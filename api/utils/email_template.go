package utils

import (
    "fmt"
    "time"
)

// BuildStyledEmail returns a simple styled HTML email.
func BuildStyledEmail(title, message, buttonText, buttonURL string) string {
    btn := ""
    if buttonText != "" && buttonURL != "" {
        btn = fmt.Sprintf(`<div style="text-align:center;"><a href="%s" class="button">%s</a></div>`, buttonURL, buttonText)
    }

    return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<title>%s - Tool Center</title>
<link href="https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;500;600;700&display=swap" rel="stylesheet">
<style>
body{margin:0;padding:0;background-color:#121212;font-family:'Poppins',sans-serif;color:#e0e0e0;}
.container{max-width:600px;margin:30px auto;background:#1e1e1e;border-radius:12px;overflow:hidden;box-shadow:0 10px 30px rgba(0,0,0,0.3);border:1px solid #333;}
.header{padding:30px;text-align:center;background:linear-gradient(135deg,#3000FF 0%%,#6200EA 100%%);}
.logo{max-width:180px;height:auto;}
.content{padding:30px;}
h1{color:#fff;font-size:28px;margin-bottom:20px;font-weight:600;text-align:center;}
p{font-size:16px;line-height:1.6;margin-bottom:20px;color:#b0b0b0;}
.button{display:inline-block;background:linear-gradient(135deg,#3000FF 0%%,#6200EA 100%%);color:#fff !important;text-decoration:none;padding:14px 28px;border-radius:8px;font-weight:600;font-size:16px;margin:25px auto;text-align:center;transition:all .3s ease;box-shadow:0 4px 15px rgba(48,0,255,.3);border:none;cursor:pointer;}
.button:hover{transform:translateY(-2px);box-shadow:0 6px 20px rgba(48,0,255,.4);}
.footer{padding:20px;text-align:center;background:#121212;border-top:1px solid #333;font-size:12px;color:#666;}
</style>
</head>
<body>
<div class="container">
<div class="header"><img src="https://tool-center.fr/assets/tc_logo.webp" alt="ToolCenter" class="logo"></div>
<div class="content">
<h1>%s</h1>
<p>%s</p>
%s
</div>
<div class="footer">© Tool Center %d. Tous droits réservés.</div>
</div>
</body>
</html>`, title, title, message, btn, time.Now().Year())
}
