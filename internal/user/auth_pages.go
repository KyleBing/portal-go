package user

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type authPageData struct {
	Title    string
	Heading  string
	Detail   string
	Year     int
	ShowForm bool
	TokenJS  template.JS
}

var authPageTmpl = template.Must(template.New("auth").Parse(authPageHTML))

func writeAuthMessage(c *gin.Context, status int, heading, detail string) {
	writeAuthPage(c, status, authPageData{
		Title:   heading + " · 日记",
		Heading: heading,
		Detail:  detail,
		Year:    time.Now().Year(),
	})
}

func writeResetForm(c *gin.Context, rawToken string) {
	tokenJSON, err := json.Marshal(rawToken)
	if err != nil {
		writeAuthMessage(c, http.StatusInternalServerError, "无法打开重置页", "请稍后重试，或重新申请重置密码。")
		return
	}
	writeAuthPage(c, http.StatusOK, authPageData{
		Title:    "重置密码 · 日记",
		Heading:  "重置密码",
		Detail:   "请设置新密码。完成后即可返回日记登录。",
		Year:     time.Now().Year(),
		ShowForm: true,
		TokenJS:  template.JS(tokenJSON),
	})
}

func writeAuthPage(c *gin.Context, status int, data authPageData) {
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	_ = authPageTmpl.Execute(c.Writer, data)
}

const authPageHTML = `<!DOCTYPE html>
<html lang="zh-Hans">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
  :root {
    --ink: #1a1a1a;
    --muted: #5a6570;
    --line: rgba(26, 26, 26, 0.12);
    --sky: #3498db;
    --ok: #27ae60;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    min-height: 100vh;
    color: var(--ink);
    font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Hiragino Sans GB", sans-serif;
    background:
      radial-gradient(800px 420px at 10% -10%, rgba(52, 152, 219, 0.18), transparent 55%),
      linear-gradient(180deg, #f0f2f5 0%, #e8ecf1 100%);
  }
  .wrap { max-width: 440px; margin: 0 auto; padding: 48px 20px 32px; }
  .brand { margin-bottom: 24px; }
  .brand strong { display: block; font-size: 18px; letter-spacing: 0.02em; }
  .card {
    background: rgba(255, 255, 255, 0.95);
    border: 1px solid var(--line);
    border-radius: 14px;
    padding: 28px 24px;
    box-shadow: 0 12px 36px rgba(26, 40, 60, 0.08);
  }
  h1 { margin: 0 0 10px; font-size: 22px; font-weight: 650; }
  .lede { margin: 0; color: var(--muted); font-size: 15px; line-height: 1.7; }
  form { margin-top: 22px; }
  label { display: block; margin: 0 0 6px; font-size: 13px; color: var(--muted); }
  input {
    width: 100%;
    margin: 0 0 14px;
    padding: 12px 14px;
    border: 1px solid var(--line);
    border-radius: 8px;
    font-size: 16px;
    background: #fff;
  }
  input:focus { outline: 2px solid rgba(52, 152, 219, 0.28); border-color: var(--sky); }
  button {
    width: 100%;
    margin-top: 4px;
    padding: 12px 16px;
    border: 0;
    border-radius: 8px;
    background: var(--sky);
    color: #fff;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
  }
  button:disabled { opacity: 0.6; cursor: default; }
  #msg { margin: 14px 0 0; font-size: 14px; line-height: 1.5; min-height: 1.2em; }
  #msg.ok { color: var(--ok); }
  #msg.err { color: #b42318; }
  .foot { margin: 28px 0 0; text-align: center; color: #8a939c; font-size: 12px; }
</style>
</head>
<body>
  <div class="wrap">
    <header class="brand">
      <strong>日记</strong>
    </header>
    <main class="card">
      <h1>{{.Heading}}</h1>
      <p class="lede">{{.Detail}}</p>
      {{if .ShowForm}}
      <form id="f">
        <label for="pw">新密码</label>
        <input type="password" id="pw" name="password" placeholder="请输入新密码" required autocomplete="new-password">
        <label for="pw2">确认密码</label>
        <input type="password" id="pw2" name="password2" placeholder="再次输入新密码" required autocomplete="new-password">
        <button type="submit">确认重置</button>
      </form>
      <p id="msg"></p>
      <script>
        const token = {{.TokenJS}};
        const form = document.getElementById('f');
        const msg = document.getElementById('msg');
        form.addEventListener('submit', async (e) => {
          e.preventDefault();
          const pw = document.getElementById('pw').value;
          const pw2 = document.getElementById('pw2').value;
          msg.className = '';
          msg.textContent = '';
          if (pw !== pw2) {
            msg.className = 'err';
            msg.textContent = '两次输入的密码不一致。';
            return;
          }
          form.querySelector('button').disabled = true;
          try {
            const r = await fetch(window.location.pathname, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ token, password: pw })
            });
            const data = await r.json().catch(() => ({}));
            if (r.ok && data.success !== false) {
              msg.className = 'ok';
              msg.textContent = data.message || '密码已更新，可以返回日记登录。';
              form.style.display = 'none';
            } else {
              msg.className = 'err';
              msg.textContent = (data && data.message) || '重置失败，链接可能已过期。请重新申请。';
              form.querySelector('button').disabled = false;
            }
          } catch {
            msg.className = 'err';
            msg.textContent = '网络错误，请稍后重试。';
            form.querySelector('button').disabled = false;
          }
        });
      </script>
      {{end}}
    </main>
    <p class="foot">© {{.Year}} Kyle · 日记</p>
  </div>
</body>
</html>
`
