package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"
)

type emailContent struct {
	LogoURL          string
	Preheader        string
	Title            string
	Intro            string
	Detail           string
	HighlightCode    string
	ActionURL        string
	ActionURLDisplay string
	ActionLabel      string
	Footer           string
	Year             int
	BrandName        string
}

var emailHTMLTmpl = template.Must(template.New("email").Parse(emailHTML))

func renderVerifyEmail(logoURL, actionURL string) (htmlBody, textBody string, err error) {
	return renderEmail(emailContent{
		LogoURL:          logoURL,
		BrandName:        "日记",
		Preheader:        "点击完成邮箱验证，24 小时内有效。",
		Title:            "验证邮箱",
		Intro:            "你好，感谢注册日记账号。",
		Detail:           "请在 24 小时内点击下方按钮完成邮箱验证，验证后即可登录使用。",
		ActionURL:        actionURL,
		ActionURLDisplay: actionURL,
		ActionLabel:      "验证邮箱",
		Footer:           "若你没有注册日记账号，请忽略本邮件。",
		Year:             time.Now().Year(),
	})
}

func renderResetEmail(logoURL, actionURL string) (htmlBody, textBody string, err error) {
	return renderEmail(emailContent{
		LogoURL:          logoURL,
		BrandName:        "日记",
		Preheader:        "点击设置新密码，1 小时内有效。",
		Title:            "重置密码",
		Intro:            "你好，我们收到了重置日记账号密码的请求。",
		Detail:           "请在 1 小时内点击下方按钮设置新密码。若不是你本人操作，请忽略本邮件，密码不会改变。",
		ActionURL:        actionURL,
		ActionURLDisplay: actionURL,
		ActionLabel:      "重置密码",
		Footer:           "若你没有申请重置密码，请忽略本邮件。",
		Year:             time.Now().Year(),
	})
}

func renderInvitationEmail(logoURL, invitationCode, registerURL, registerURLDisplay string) (htmlBody, textBody string, err error) {
	return renderEmail(emailContent{
		LogoURL:          logoURL,
		BrandName:        "日记",
		Preheader:        "你的日记注册邀请码已下发。",
		Title:            "注册邀请码",
		Intro:            "你好，你的付费注册申请已通过审核。",
		Detail:           "请使用申请邮箱与下方邀请码完成注册；注册后请查收验证邮件并完成验证再登录。也可直接点击下方按钮，系统会为你预填信息。",
		HighlightCode:    invitationCode,
		ActionURL:        registerURL,
		ActionURLDisplay: registerURLDisplay,
		ActionLabel:      "前往注册",
		Footer:           "若你没有提交过注册申请，请忽略本邮件。",
		Year:             time.Now().Year(),
	})
}

const supportEmail = "kylebing@163.com"

func renderRegistrationApplicationSubmittedEmail(logoURL string) (htmlBody, textBody string, err error) {
	return renderEmail(emailContent{
		LogoURL:          logoURL,
		BrandName:        "日记",
		Preheader:        "你的日记注册申请已收到，请等待审核。",
		Title:            "注册申请已提交",
		Intro:            "欢迎使用日记！你的付费注册申请已成功提交。",
		Detail:           "我们会在 1 日内完成审核，审核结果将发送至本邮箱；通过后会附邀请码，请使用申请邮箱完成注册并验证后再登录。",
		ActionURL:        "mailto:" + supportEmail,
		ActionURLDisplay: supportEmail,
		ActionLabel:      "邮件联系支持",
		Footer:           "有任何问题欢迎邮件联系 " + supportEmail + "。若你没有提交过注册申请，请忽略本邮件。",
		Year:             time.Now().Year(),
	})
}

func renderEmail(c emailContent) (htmlBody, textBody string, err error) {
	if c.ActionURLDisplay == "" {
		c.ActionURLDisplay = c.ActionURL
	}
	var buf bytes.Buffer
	if err := emailHTMLTmpl.Execute(&buf, c); err != nil {
		return "", "", err
	}
	detail := c.Detail
	if c.HighlightCode != "" {
		detail = detail + "\n\n邀请码：\n" + c.HighlightCode
	}
	var actionBlock string
	if strings.TrimSpace(c.ActionURL) != "" && strings.TrimSpace(c.ActionLabel) != "" {
		linkLine := c.ActionURLDisplay
		if c.ActionURLDisplay != c.ActionURL {
			linkLine = c.ActionURLDisplay + "\n（实际链接已包含预填信息）"
		}
		actionBlock = fmt.Sprintf("\n\n%s\n%s", c.ActionLabel, linkLine)
	}
	text := fmt.Sprintf("%s\n\n%s\n%s%s\n\n%s\n\n© %d Kyle · %s",
		c.Title, c.Intro, detail, actionBlock, c.Footer, c.Year, c.BrandName)
	return buf.String(), strings.TrimSpace(text), nil
}

// 邮件客户端兼容的表格布局，仅用内联样式。
const emailHTML = `<!DOCTYPE html>
<html lang="zh-Hans">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="x-ua-compatible" content="ie=edge">
<title>{{.Title}} · {{.BrandName}}</title>
</head>
<body style="margin:0;padding:0;background-color:#f0f2f5;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;color:#f0f2f5;font-size:1px;line-height:1px;">{{.Preheader}}</div>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#f0f2f5;">
  <tr>
    <td align="center" style="padding:32px 16px;">
      <table role="presentation" width="560" cellpadding="0" cellspacing="0" border="0" style="max-width:560px;width:100%;background-color:#ffffff;border-radius:12px;overflow:hidden;">
        <tr>
          <td bgcolor="#2c3e50" style="background-color:#2c3e50;padding:20px 28px;">
            <table role="presentation" cellpadding="0" cellspacing="0" border="0">
              <tr>
                {{if .LogoURL}}
                <td valign="middle" style="padding-right:12px;">
                  <img src="{{.LogoURL}}" width="40" height="40" alt="{{.BrandName}}" style="display:block;width:40px;height:40px;border:0;border-radius:10px;">
                </td>
                {{end}}
                <td valign="middle">
                  <div style="font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Hiragino Sans GB',sans-serif;color:#ffffff;font-size:17px;font-weight:600;line-height:1.2;">{{.BrandName}}</div>
                </td>
              </tr>
            </table>
          </td>
        </tr>
        <tr>
          <td height="4" bgcolor="#3498db" style="background-color:#3498db;font-size:0;line-height:0;height:4px;">&nbsp;</td>
        </tr>
        <tr>
          <td style="padding:32px 28px 8px;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC','Hiragino Sans GB',sans-serif;color:#1a1a1a;">
            <h1 style="margin:0 0 16px;font-size:22px;line-height:1.3;font-weight:650;">{{.Title}}</h1>
            <p style="margin:0 0 12px;color:#5a6570;font-size:15px;line-height:1.7;">{{.Intro}}</p>
            <p style="margin:0 0 20px;color:#5a6570;font-size:15px;line-height:1.7;">{{.Detail}}</p>
            {{if .HighlightCode}}
            <p style="margin:0 0 8px;color:#5a6570;font-size:14px;line-height:1.5;">邀请码</p>
            <p style="margin:0 0 28px;padding:14px 16px;background-color:#f4f6f8;border-radius:8px;font-family:ui-monospace,Menlo,Consolas,monospace;font-size:20px;font-weight:700;letter-spacing:0.5px;color:#1a1a1a;line-height:1.4;word-break:break-all;">{{.HighlightCode}}</p>
            {{end}}
            {{if and .ActionURL .ActionLabel}}
            <table role="presentation" cellpadding="0" cellspacing="0" border="0">
              <tr>
                <td bgcolor="#3498db" style="background-color:#3498db;border-radius:8px;">
                  <a href="{{.ActionURL}}" style="display:inline-block;padding:13px 28px;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC',sans-serif;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;">{{.ActionLabel}}</a>
                </td>
              </tr>
            </table>
            <p style="margin:28px 0 0;color:#8a939c;font-size:12px;line-height:1.65;">如果按钮无法打开，请将以下链接复制到浏览器：<br>
              <a href="{{.ActionURL}}" style="color:#3498db;word-break:break-all;text-decoration:underline;">{{.ActionURLDisplay}}</a>
            </p>
            {{end}}
          </td>
        </tr>
        <tr>
          <td style="padding:20px 28px 28px;border-top:1px solid #e8ecef;font-family:-apple-system,BlinkMacSystemFont,'PingFang SC',sans-serif;color:#8a939c;font-size:12px;line-height:1.7;">
            {{.Footer}}<br>
            这是一封自动发送的邮件，请勿直接回复。<br>
            © {{.Year}} Kyle · {{.BrandName}}
          </td>
        </tr>
      </table>
    </td>
  </tr>
</table>
</body>
</html>
`
