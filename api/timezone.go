package api

type TimezoneItem struct {
	Value string
	Label string
}

var TimezonesItems = []TimezoneItem{
	{"Europe/Berlin", "(UTC+01:00) Amsterdam, Berlin, Bern, Rome, Stockholm, Vienna"},
	{"Europe/Belgrade", "(UTC+01:00) Belgrade, Bratislava, Budapest, Ljubljana, Prague"},
	{"Europe/Brussels", "(UTC+01:00) Brussels, Copenhagen, Madrid, Paris"},
	{"Europe/Warsaw", "(UTC+01:00) Sarajevo, Skopje, Warsaw, Zagreb"},
	{"Africa/Lagos", "(UTC+01:00) West Central Africa"},
	{"Africa/Windhoek", "(UTC+01:00) Windhoek"},
	{"Europe/Athens", "(UTC+02:00) Athens, Bucharest"},
	{"Asia/Beirut", "(UTC+02:00) Beirut"},
	{"Africa/Cairo", "(UTC+02:00) Cairo"},
	{"Asia/Damascus", "(UTC+02:00) Damascus"},
	{"Europe/Bucharest", "(UTC+02:00) E. Europe"},
	{"Africa/Harare", "(UTC+02:00) Harare, Pretoria"},
	{"Europe/Helsinki", "(UTC+02:00) Helsinki, Kyiv, Riga, Sofia, Tallinn, Vilnius"},
	{"Europe/Istanbul", "(UTC+03:00) Istanbul"},
	{"Asia/Jerusalem", "(UTC+02:00) Jerusalem"},
	{"Africa/Tripoli", "(UTC+02:00) Tripoli"},
	{"Asia/Amman", "(UTC+03:00) Amman"},
	{"Asia/Baghdad", "(UTC+03:00) Baghdad"},
	{"Europe/Kaliningrad", "(UTC+02:00) Kaliningrad"},
	{"Asia/Riyadh", "(UTC+03:00) Kuwait, Riyadh"},
	{"Africa/Nairobi", "(UTC+03:00) Nairobi"},
	{"Europe/Moscow", "(UTC+03:00) Moscow, St. Petersburg, Volgograd, Minsk"},
	{"Europe/Samara", "(UTC+04:00) Samara, Ulyanovsk, Saratov"},
	{"Asia/Tehran", "(UTC+03:30) Tehran"},
	{"Asia/Dubai", "(UTC+04:00) Abu Dhabi, Muscat"},
	{"Asia/Baku", "(UTC+04:00) Baku"},
	{"Indian/Mauritius", "(UTC+04:00) Port Louis"},
	{"Asia/Tbilisi", "(UTC+04:00) Tbilisi"},
	{"Asia/Yerevan", "(UTC+04:00) Yerevan"},
	{"Asia/Kabul", "(UTC+04:30) Kabul"},
	{"Asia/Tashkent", "(UTC+05:00) Ashgabat, Tashkent"},
	{"Asia/Yekaterinburg", "(UTC+05:00) Yekaterinburg"},
	{"Asia/Karachi", "(UTC+05:00) Islamabad, Karachi"},
	{"Asia/Kolkata", "(UTC+05:30) Chennai, Kolkata, Mumbai, New Delhi"},
	{"Asia/Colombo", "(UTC+05:30) Sri Jayawardenepura"},
	{"Asia/Kathmandu", "(UTC+05:45) Kathmandu"},
	{"Asia/Almaty", "(UTC+06:00) Nur-Sultan (Astana)"},
	{"Asia/Dhaka", "(UTC+06:00) Dhaka"},
	{"Asia/Yangon", "(UTC+06:30) Yangon (Rangoon)"},
	{"Asia/Bangkok", "(UTC+07:00) Bangkok, Hanoi, Jakarta"},
	{"Asia/Novosibirsk", "(UTC+07:00) Novosibirsk"},
	{"Asia/Shanghai", "(UTC+08:00) Beijing, Chongqing, Hong Kong, Urumqi"},
	{"Asia/Krasnoyarsk", "(UTC+08:00) Krasnoyarsk"},
	{"Asia/Singapore", "(UTC+08:00) Kuala Lumpur, Singapore"},
	{"Australia/Perth", "(UTC+08:00) Perth"},
	{"Asia/Taipei", "(UTC+08:00) Taipei"},
	{"Asia/Ulaanbaatar", "(UTC+08:00) Ulaanbaatar"},
	{"Asia/Irkutsk", "(UTC+08:00) Irkutsk"},
	{"Asia/Tokyo", "(UTC+09:00) Osaka, Sapporo, Tokyo"},
	{"Asia/Seoul", "(UTC+09:00) Seoul"},
	{"Australia/Adelaide", "(UTC+09:30) Adelaide"},
	{"Australia/Darwin", "(UTC+09:30) Darwin"},
	{"Australia/Brisbane", "(UTC+10:00) Brisbane"},
	{"Australia/Sydney", "(UTC+10:00) Canberra, Melbourne, Sydney"},
	{"Pacific/Port_Moresby", "(UTC+10:00) Guam, Port Moresby"},
	{"Australia/Hobart", "(UTC+10:00) Hobart"},
	{"Asia/Yakutsk", "(UTC+09:00) Yakutsk"},
	{"Pacific/Guadalcanal", "(UTC+11:00) Solomon Is., New Caledonia"},
	{"Asia/Vladivostok", "(UTC+11:00) Vladivostok"},
	{"Pacific/Auckland", "(UTC+12:00) Auckland, Wellington"},
	{"Etc/GMT-12", "(UTC+12:00) Coordinated Universal Time+12"},
	{"Pacific/Fiji", "(UTC+12:00) Fiji"},
	{"Asia/Magadan", "(UTC+12:00) Magadan"},
	{"Asia/Kamchatka", "(UTC+12:00) Petropavlovsk-Kamchatsky - Old"},
	{"Pacific/Tongatapu", "(UTC+13:00) Nuku`alofa"},
	{"Pacific/Apia", "(UTC+13:00) Samoa"},
	{"Etc/GMT+12", "(UTC-12:00) International Date Line West"},
	{"Etc/GMT+11", "(UTC-11:00) Coordinated Universal Time-11"},
	{"Pacific/Honolulu", "(UTC-10:00) Hawaii"},
	{"America/Anchorage", "(UTC-09:00) Alaska"},
	{"America/Tijuana", "(UTC-08:00) Baja California"},
	{"America/Los_Angeles", "(UTC-07:00) Pacific Time (US & Canada)"},
	{"America/Los_Angeles", "(UTC-08:00) Pacific Time (US & Canada)"},
	{"America/Phoenix", "(UTC-07:00) Arizona"},
	{"America/Chihuahua", "(UTC-07:00) Chihuahua, La Paz, Mazatlan"},
	{"America/Denver", "(UTC-07:00) Mountain Time (US & Canada)"},
	{"America/Guatemala", "(UTC-06:00) Central America"},
	{"America/Chicago", "(UTC-06:00) Central Time (US & Canada)"},
	{"America/Mexico_City", "(UTC-06:00) Guadalajara, Mexico City, Monterrey"},
	{"America/Regina", "(UTC-06:00) Saskatchewan"},
	{"America/Bogota", "(UTC-05:00) Bogota, Lima, Quito"},
	{"America/New_York", "(UTC-05:00) Eastern Time (US & Canada)"},
	{"America/Indiana/Indianapolis", "(UTC-05:00) Indiana (East)"},
	{"America/Caracas", "(UTC-04:30) Caracas"},
	{"America/Asuncion", "(UTC-04:00) Asuncion"},
	{"America/Halifax", "(UTC-04:00) Atlantic Time (Canada)"},
	{"America/Cuiaba", "(UTC-04:00) Cuiaba"},
	{"America/La_Paz", "(UTC-04:00) Georgetown, La Paz, Manaus, San Juan"},
	{"America/Santiago", "(UTC-04:00) Santiago"},
	{"America/St_Johns", "(UTC-03:30) Newfoundland"},
	{"America/Sao_Paulo", "(UTC-03:00) Brasilia"},
	{"America/Argentina/Buenos_Aires", "(UTC-03:00) Buenos Aires"},
	{"America/Cayenne", "(UTC-03:00) Cayenne, Fortaleza"},
	{"America/Godthab", "(UTC-03:00) Greenland"},
	{"America/Montevideo", "(UTC-03:00) Montevideo"},
	{"America/Bahia", "(UTC-03:00) Salvador"},
	{"Etc/GMT+2", "(UTC-02:00) Coordinated Universal Time-02"},
	{"Etc/GMT+2", "(UTC-02:00) Mid-Atlantic - Old"},
	{"Atlantic/Azores", "(UTC-01:00) Azores"},
	{"Atlantic/Cape_Verde", "(UTC-01:00) Cape Verde Is."},
	{"Africa/Casablanca", "(UTC) Casablanca"},
	{"Etc/UTC", "(UTC) Coordinated Universal Time"},
	{"Europe/London", "(UTC) Edinburgh, London"},
	{"Europe/London", "(UTC+01:00) Edinburgh, London"},
	{"Europe/Dublin", "(UTC) Dublin, Lisbon"},
	{"Africa/Monrovia", "(UTC) Monrovia, Reykjavik"},
}

func GetTimezones() []string {
	timezones := make([]string, len(TimezonesItems))
	for i, item := range TimezonesItems {
		timezones[i] = item.Value
	}
	return timezones
}
