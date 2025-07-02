document.addEventListener('keydown', function(e) {
    if (e.key === 'F12') {
      triggerConsoleMessage();
    }
    if (e.ctrlKey && e.shiftKey && (e.key === 'I' || e.key === 'J')) {
      triggerConsoleMessage();
    }
    if (e.ctrlKey && e.key === 'u') {
      triggerConsoleMessage();
    }
  });
  
  function triggerConsoleMessage() {
      console.clear();
      console.log(`%c
                                    @@@@@@
                              @@@@@@@@@@@@@@@@@@
                           @@@@@@@@@@@@@@@@@@@@@@@@
                         @@@@@@@@@@@@@@@@@@@@@@@@@@@@
                       @@@@@@@@@@
                      @@@@@@@@@    @@@@@@@@@@                  @@@@@@
                      @@@@@@@@          @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
                      @@@@@@@@           @@@@@@@@@@@@@@@@@@@@@@@    @@@
                      @@@@@@@@@         @@@@@                 @@@@@@@@
                       @@@@@@@@@   @@@@@@@@@
                        @@@@@@@@@          @@@@@@@@@@@
                          @@@@@@@@@@@@@@@@@@@@@@@@@@
                            @@@@@@@@@@@@@@@@@@@@@
                                 @@@@@@@@@@@
      `, 'color: white; font-family: monospace; font-size: 10px; line-height: 1.2');
      console.log('%c🚨 ATTENTION !!! ⚠️', 'color: red; font-size: 28px; font-weight: bold; padding: 10px 0;');
      console.log('%c🛡️ Si quelqu’un te demande d’aller dans les performances ou le "local storage", n’y va surtout pas ! Il pourrait voler ton compte.', 
                  'color: red; font-size: 16px; font-weight: 600; line-height: 1.5;');
      console.log('%c━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━', 
                  'color: gray; font-size: 12px; font-family: monospace;');
      console.log('%c💎 Tu sais ce que tu fais ? Alors pourquoi tu bosses pas avec nous ?\n📩 Contacte-nous à : gabex@gabex.xyz', 
                  'color: #3000FF; font-size: 18px; font-weight: bold; line-height: 1.5;');
      }