package configgen

import (
	"strings"
	"testing"
	"text/template"
)

func render(t *testing.T, serverParams map[string]string) string {
	t.Helper()
	tpl, err := template.ParseFS(awgTemplateFS, "templates/awg/client.conf")
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]string{
		"ClientAddr":         "10.8.1.30/32",
		"ClientPrivKey":      "ZonoIPpufcnr4ep2J2yo0pGafx3kRH8o5LM8fIFmaVc=",
		"ServerPubKey":       "bNSu7A5/MWRciCIEznjcTWj8e0WTQjGOSEkGS9QMcF0=",
		"ClientPresharedKey": "YFvGmCpOy05fSuiwZYP40A/vxctulgdqgGJ2psEcm5k=",
		"ServerAddr":         "service4.shpagafoxthesecond.com",
		"ServerPort":         "36623",
	}
	for _, k := range awgMirroredParams {
		data[k] = serverParams[k]
	}
	for k, def := range awgTimingDefaults {
		if v, ok := serverParams[k]; ok {
			data[k] = v
			continue
		}
		data[k] = def
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		t.Fatal(err)
	}
	return sb.String()
}

func TestRenderFull(t *testing.T) {
	t.Log("\n" + render(t, map[string]string{
		"Jc": "4", "Jmin": "10", "Jmax": "50",
		"S1": "100", "S2": "124", "S3": "45", "S4": "12",
		"H1": "1", "H2": "2", "H3": "3", "H4": "4",
		"HeaderProtectionKey":    "MHjJAEAAAj/FJndvbWsRBy95hXJPQwXhuhcUB0tMT48=",
		"ContentPaddingAddition": "10-100",
		"RandomTrailers":         "on",
		"DisableCookies":         "on",
	}))
}

func TestRenderLegacyServer(t *testing.T) {
	t.Log("\n" + render(t, map[string]string{
		"Jc": "4", "Jmin": "10", "Jmax": "50",
		"S1": "100", "S2": "124",
		"H1": "1", "H2": "2", "H3": "3", "H4": "4",
	}))
}
