package models

import "errors"

type Currency string

func (c Currency) Validate() error {
	// Ref: http://country.io/currency.json
	allCurrencies := map[Currency]string{"AED": "AE", "AFN": "AF", "ALL": "AL", "AMD": "AM", "ANG": "SX", "AOA": "AO", "ARS": "AR", "AUD": "CX", "AWG": "AW", "AZN": "AZ", "BAM": "BA", "BBD": "BB", "BDT": "BD", "BGN": "BG", "BHD": "BH", "BIF": "BI", "BMD": "BM", "BND": "BN", "BOB": "BO", "BRL": "BR", "BSD": "BS", "BTN": "BT", "BWP": "BW", "BYR": "BY", "BZD": "BZ", "CAD": "CA", "CDF": "CD", "CHF": "CH", "CLP": "CL", "CNY": "CN", "COP": "CO", "CRC": "CR", "CUP": "CU", "CVE": "CV", "CZK": "CZ", "DJF": "DJ", "DKK": "FO", "DOP": "DO", "DZD": "DZ", "EGP": "EG", "ERN": "ER", "ETB": "ET", "EUR": "VA", "FJD": "FJ", "FKP": "FK", "GBP": "GS", "GEL": "GE", "GHS": "GH", "GIP": "GI", "GMD": "GM", "GNF": "GN", "GTQ": "GT", "GYD": "GY", "HKD": "HK", "HNL": "HN", "HRK": "HR", "HTG": "HT", "HUF": "HU", "IDR": "ID", "ILS": "IL", "INR": "IN", "IQD": "IQ", "IRR": "IR", "ISK": "IS", "JMD": "JM", "JOD": "JO", "JPY": "JP", "KES": "KE", "KGS": "KG", "KHR": "KH", "KMF": "KM", "KPW": "KP", "KRW": "KR", "KWD": "KW", "KYD": "KY", "KZT": "KZ", "LAK": "LA", "LBP": "LB", "LKR": "LK", "LRD": "LR", "LSL": "LS", "LTL": "LT", "LYD": "LY", "MAD": "MA", "MDL": "MD", "MGA": "MG", "MKD": "MK", "MMK": "MM", "MNT": "MN", "MOP": "MO", "MRO": "MR", "MUR": "MU", "MVR": "MV", "MWK": "MW", "MXN": "MX", "MYR": "MY", "MZN": "MZ", "NAD": "NA", "NGN": "NG", "NIO": "NI", "NOK": "BV", "NPR": "NP", "NZD": "NU", "OMR": "OM", "PAB": "PA", "PEN": "PE", "PGK": "PG", "PHP": "PH", "PKR": "PK", "PLN": "PL", "PYG": "PY", "QAR": "QA", "RON": "RO", "RSD": "RS", "RUB": "RU", "RWF": "RW", "SAR": "SA", "SBD": "SB", "SCR": "SC", "SDG": "SD", "SEK": "SE", "SGD": "SG", "SHP": "SH", "SLL": "SL", "SOS": "SO", "SRD": "SR", "SSP": "SS", "STD": "ST", "SYP": "SY", "SZL": "SZ", "THB": "TH", "TJS": "TJ", "TMT": "TM", "TND": "TN", "TOP": "TO", "TRY": "TR", "TTD": "TT", "TWD": "TW", "TZS": "TZ", "UAH": "UA", "UGX": "UG", "USD": "TL", "UYU": "UY", "UZS": "UZ", "VEF": "VE", "VND": "VN", "VUV": "VU", "WST": "WS", "XAF": "CF", "XCD": "GD", "XOF": "NE", "XPF": "PF", "YER": "YE", "ZAR": "ZA", "ZMK": "ZM", "ZWL": "ZW"}
	if _, ok := allCurrencies[c]; !ok {
		return errors.New("invalid currency")
	}

	return nil
}
