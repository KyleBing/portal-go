package weather

import "testing"

func TestOfficialPathFromRequest(t *testing.T) {
	cases := map[string]string{
		"/portal/v7/weather/24h": "/v7/weather/24h",
		"/v7/weather/now":        "/v7/weather/now",
		"/portal/geo/v2/city/lookup": "/geo/v2/city/lookup",
		"/portal/weather/v1/hourly/39.92/116.41": "/weather/v1/hourly/39.92/116.41",
	}
	for in, want := range cases {
		if got := officialPathFromRequest(in); got != want {
			t.Errorf("officialPathFromRequest(%q)=%q want %q", in, got, want)
		}
	}
}

func TestIsOfficialQWeatherPath(t *testing.T) {
	ok := []string{
		"/v7/weather/24h",
		"/v7/weather/now",
		"/v7/minutely/5m",
		"/v7/warning/now",
		"/v7/indices/1d",
		"/v7/air/now",
		"/v7/astronomy/sun",
		"/v7/tropical/storm-list",
		"/v7/ocean/tide",
		"/v7/historical/weather",
		"/geo/v2/city/lookup",
		"/weather/v1/hourly/39.92/116.41",
		"/airquality/v1/current/39.92/116.41",
		"/weatheralert/v1/current/39.92/116.41",
		"/solarradiation/v1/forecast/39.92/116.41",
	}
	for _, p := range ok {
		if !isOfficialQWeatherPath(p) {
			t.Errorf("want allow %q", p)
		}
	}
	deny := []string{
		"/weather/hourly",
		"/weather/now",
		"/v7/../secret",
		"https://evil.example/v7/weather/now",
		"/user/login",
		"/v7/weather/now/../../user",
		"/v8/weather/now",
	}
	for _, p := range deny {
		got := officialPathFromRequest(p)
		if isOfficialQWeatherPath(got) {
			t.Errorf("want deny %q (got %q)", p, got)
		}
	}
}
