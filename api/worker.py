import json, pathlib, time, pymysql, smtplib, email.message, datetime, sys

root = pathlib.Path(__file__).resolve().parent
try:
    with open(root / 'config.json') as f:
        cfg = json.load(f)
except Exception as e:
    print(f"Erreur ouverture config.json: {e}")
    sys.exit(1)

db = cfg['database']
mail = cfg['email']

try:
    conn = pymysql.connect(host=db['host'], port=db['port'], user=db['user'],
                           password=db['password'], database=db['dbname'], autocommit=True)
except Exception as e:
    print(f"Erreur connexion BDD: {e}")
    sys.exit(1)

def _html(sub, body):
    body_html = body.replace('\n', '<br>')
    year = datetime.datetime.now().year
    return f"""<!DOCTYPE html><html lang='fr'><head><meta charset='UTF-8'><title>{sub}</title>
    <link href='https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;500;600;700&display=swap' rel='stylesheet'>
    </head><body style='margin:0;padding:0;background:#121212;font-family:Poppins,Arial,sans-serif;color:#e0e0e0;'>
      <div style='max-width:600px;margin:30px auto;background:#1e1e1e;border-radius:12px;overflow:hidden;
                  box-shadow:0 10px 30px rgba(0,0,0,.3);border:1px solid #333;'>
        <div style='padding:30px;text-align:center;
                    background:linear-gradient(135deg,#3000FF 0%,#6200EA 100%);'>
          <img src='https://tool-center.fr/assets/tc_logo.webp' alt='Tool Center Logo'
               style='border-radius:9999px;max-width:180px;height:auto;display:block;margin:auto;'>
        </div>
        <div style='padding:30px;'>
          {body_html}
        </div>
        <div style='padding:20px;text-align:center;background:#121212;border-top:1px solid #333;
                    font-size:12px;color:#666;'>© Tool Center {year}. Tous droits réservés.</div>
      </div></body></html>"""

def send(to_addr, sub, body):
    try:
        msg = email.message.EmailMessage()
        msg['From'] = mail['username']
        msg['To'] = to_addr
        msg['Subject'] = sub
        msg.set_content(body)
        msg.add_alternative(_html(sub, body), subtype='html')
        with smtplib.SMTP_SSL(mail['host'], mail['port']) as s:
            s.login(mail['username'], mail['password'])
            s.send_message(msg)
    except Exception as e:
        print(f"Erreur envoi mail à {to_addr}: {e}")

while True:
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT user_id,email FROM users WHERE email_verified_at IS NULL AND created_at < NOW() - INTERVAL 10 MINUTE")
            for uid, adr in cur.fetchall():
                send(adr, 'Compte supprimé', f'Ton compte #{uid} a été supprimé faute de vérification.')
                cur.execute("DELETE FROM users WHERE user_id=%s", (uid,))
            cur.execute("SELECT queue_id,to_email,subject,body FROM email_queue")
            for qid, to_addr, sub, body in cur.fetchall():
                send(to_addr, sub, body)
                cur.execute("DELETE FROM email_queue WHERE queue_id=%s", (qid,))
    except Exception as e:
        print(f"Erreur traitement BDD: {e}")
    time.sleep(600)
