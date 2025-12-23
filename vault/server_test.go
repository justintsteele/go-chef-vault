package vault

import (
	"net/http"
	"net/http/httptest"

	"github.com/go-chef/chef"
)

type keyPair struct {
	private,
	public,
	kind string
}

const (
	userid     = "tester"
	requestURL = "http://localhost:80"

	// Generated from
	// openssl genrsa -out privkey.pem 2048
	// perl -pe 's/\n/\\n/g' privkey.pem
	privateKeyPKCS1 = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAx12nDxxOwSPHRSJEDz67a0folBqElzlu2oGMiUTS+dqtj3FU
h5lJc1MjcprRVxcDVwhsSSo9948XEkk39IdblUCLohucqNMzOnIcdZn8zblN7Cnp
W03UwRM0iWX1HuwHnGvm6PKeqKGqplyIXYO0qlDWCzC+VaxFTwOUk31MfOHJQn4y
fTrfuE7h3FTElLBu065SFp3dPICIEmWCl9DadnxbnZ8ASxYQ9xG7hmZduDgjNW5l
3x6/EFkpym+//D6AbWDcVJ1ovCsJL3CfH/NZC3ekeJ/aEeLxP/vaCSH1VYC5VsYK
5Qg7SIa6Nth3+RZz1hYOoBJulEzwljznwoZYRQIDAQABAoIBADPQol+qAsnty5er
PTcdHcbXLJp5feZz1dzSeL0gdxja/erfEJIhg9aGUBs0I55X69VN6h7l7K8PsHZf
MzzJhUL4QJJETOYP5iuVhtIF0I+DTr5Hck/5nYcEv83KAvgjbiL4ZE486IF5awnL
2OE9HtJ5KfhEleNcX7MWgiIHGb8G1jCqu/tH0GI8Z4cNgUrXMbczGwfbN/5Wc0zo
Dtpe0Tec/Fd0DLFwRiAuheakPjlVWb7AGMDX4TyzCXfMpS1ul2jk6nGFk77uQozF
PQUawCRp+mVS4qecgq/WqfTZZbBlW2L18/kpafvsxG8kJ7OREtrb0SloZNFHEc2Q
70GbgKECgYEA6c/eOrI3Uour1gKezEBFmFKFH6YS/NZNpcSG5PcoqF6AVJwXg574
Qy6RatC47e92be2TT1Oyplntj4vkZ3REv81yfz/tuXmtG0AylH7REbxubxAgYmUT
18wUAL4s3TST2AlK4R29KwBadwUAJeOLNW+Rc4xht1galsqQRb4pUzkCgYEA2kj2
vUhKAB7QFCPST45/5q+AATut8WeHnI+t1UaiZoK41Jre8TwlYqUgcJ16Q0H6KIbJ
jlEZAu0IsJxjQxkD4oJgv8n5PFXdc14HcSQ512FmgCGNwtDY/AT7SQP3kOj0Rydg
N02uuRb/55NJ07Bh+yTQNGA+M5SSnUyaRPIAMW0CgYBgVU7grDDzB60C/g1jZk/G
VKmYwposJjfTxsc1a0gLJvSE59MgXc04EOXFNr4a+oC3Bh2dn4SJ2Z9xd1fh8Bur
UwCLwVE3DBTwl2C/ogiN4C83/1L4d2DXlrPfInvloBYR+rIpUlFweDLNuve2pKvk
llU9YGeaXOiHnGoY8iKgsQKBgQDZKMOHtZYhHoZlsul0ylCGAEz5bRT0V8n7QJlw
12+TSjN1F4n6Npr+00Y9ov1SUh38GXQFiLq4RXZitYKu6wEJZCm6Q8YXd1jzgDUp
IyAEHNsrV7Y/fSSRPKd9kVvGp2r2Kr825aqQasg16zsERbKEdrBHmwPmrsVZhi7n
rlXw1QKBgQDBOyUJKQOgDE2u9EHybhCIbfowyIE22qn9a3WjQgfxFJ+aAL9Bg124
fJIEzz43fJ91fe5lTOgyMF5TtU5ClAOPGtlWnXU0e5j3L4LjbcqzEbeyxvP3sn1z
dYkX7NdNQ5E6tcJZuJCGq0HxIAQeKPf3x9DRKzMnLply6BEzyuAC4g==
-----END RSA PRIVATE KEY-----
`
	// Generated from
	// openssl rsa -in privkey.pem -pubout -out pubkey.pem
	// perl -pe 's/\n/\\n/g' pubkey.pem
	publicKeyPKCS1 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAx12nDxxOwSPHRSJEDz67
a0folBqElzlu2oGMiUTS+dqtj3FUh5lJc1MjcprRVxcDVwhsSSo9948XEkk39Idb
lUCLohucqNMzOnIcdZn8zblN7CnpW03UwRM0iWX1HuwHnGvm6PKeqKGqplyIXYO0
qlDWCzC+VaxFTwOUk31MfOHJQn4yfTrfuE7h3FTElLBu065SFp3dPICIEmWCl9Da
dnxbnZ8ASxYQ9xG7hmZduDgjNW5l3x6/EFkpym+//D6AbWDcVJ1ovCsJL3CfH/NZ
C3ekeJ/aEeLxP/vaCSH1VYC5VsYK5Qg7SIa6Nth3+RZz1hYOoBJulEzwljznwoZY
RQIDAQAB
-----END PUBLIC KEY-----
`
	// Generated from
	// openssl dsaparam -out dsaparam.pem 2048
	// openssl gendsa  -out privkey.pem dsaparam.pem
	// perl -pe 's/\n/\\n/g' privkey.pem
	badPrivateKeyPKCS1 = `
-----BEGIN DSA PRIVATE KEY-----
MIIDVgIBAAKCAQEApv0SsaKRWyn0IrbI6i547c/gldLQ3vB5xoSuTkVOvmD3HfuE
EVPKMS+XKlhgHOJy677zYNKUOIR78vfDVr1M89w19NSic81UwGGaOkrjQWOkoHaA
BS4046AzYKWqHWQNn9dm7WdQlbMBcBv9u+J6EqlzstPwWVaRdbAzyPtwQZRF5WfC
OcrQr8XpXbKsPh55FzfvFpu4KEKTY+8ynLz9uDNW2iAxj9NtRlUHQNqKQvjQsr/8
4pVrEBh+CnzNrmPXQIbyxV0y8WukAo3I3ZXK5nsUcJhFoVCRx4aBlp9W96mYZ7OE
dPCkFsoVhUNFo0jlJhMPODR1NXy77c4v1Kh6xwIhAJwFm6CQBOWJxZdGo2luqExE
acUG9Hkr2qd0yccgs2tFAoIBAQCQJCwASD7X9l7nZyZvJpXMe6YreGaP3VbbHCz8
GHs1P5exOausfJXa9gRLx2qDW0sa1ZyFUDnd2Dt810tgAhY143lufNoV3a4IRHpS
Fm8jjDRMyBQ/BrLBBXgpwiZ9LHBuUSeoRKY0BdyRsULmcq2OaBq9J38NUblWSe2R
NjQ45X6SGgUdHy3CrQtLjCA9l8+VPg3l05IBbXIhVSllP5AUmMG4T9x6M7NHEoSr
c7ewKSJNvc1C8+G66Kfz8xcChKcKC2z1YzvxrlcDHF+BBLw1Ppp+yMBfhQDWIZfe
6tpiKEEyWoyi4GkzQ+vooFIriaaL+Nnggh+iJ7BEUByHBaHnAoIBAFUxSB3bpbbp
Vna0HN6b+svuTCFhYi9AcmI1dcyEFKycUvZjP/X07HvX2yrL8aGxMJgF6RzPob/F
+SZar3u9Fd8DUYLxis6/B5d/ih7GnfPdChrDOJM1nwlferTGHXd1TBDzugpAovCe
JAjXiPsGmcCi9RNyoGib/FgniT7IKA7s3yJAzYSeW3wtLToSNGFJHn+TzFDBuWV4
KH70bpEV84JIzWo0ejKzgMBQ0Zrjcsm4lGBtzaBqGSvOrlIVFuSWFYUxrSTTxthQ
/JYz4ch8+HsQC/0HBuJ48yALDCVKsWq4Y21LRRJIOC25DfjwEYWWaKNGlDDsJA1m
Y5WF0OX+ABcCIEXhrzI1NddyFwLnfDCQ+sy6HT8/xLKXfaipd2rpn3gL
-----END DSA PRIVATE KEY-----
`
	// Generated from
	// openssl genpkey -out rsakey.pem -algorithm RSA -pkeyopt rsa_keygen_bits:2048
	// openssl genrsa -out privkey.pem 2048
	privateKeyPKCS8 = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDNjtxSUP5FjiD9
a0KXByeLPE1y5d7G1WpJOo6YgAJjFUFPYs8+EtF7MzWpxvcRQEuYgrR7K5E7ZmSk
uM3fg+kWessqrc8qZLx3LFVv7C2O2IT0s2riHjBbBOjLbM0Ps9uX5u5vgyIOlEGz
o1dw5AMDi52QjjfROMML7WqRLMY7jcRuK7IpL5UhnAtKnOrakHSzxMHqIC2ZQnsJ
Es2Rnj7ihgr6VZ66FEEUcIqbUwZDEHYsamkg4bCFHB+P925FeZfQtBDBGlFGeNSs
mDOKrw66I2wDdq/BZ7MN3y/tdpda0H+95qYRye2FeyL9uSoREWaAv5PemQYGt2wc
xmkNoImRAgMBAAECggEABFJ2q3xsfEXqx6lTsx1BZZoU/s96ia+/Fl8W1HoMkszF
nMe1F9cJdI+1FybJ1yEE9eX5qYVW/mq+vv/rxEFfy0s1rmYNLxUDKXZTLZFHu/Mt
iH+lRa/g0GkgA/b7sNLVUTJX3RxiwO+5Ge/bTNJehdqPq5Rx9AI/h6asUPUiDep5
gy22eGh8hNYXrDvZxQBe8stVw11PSItn5pgYTtlLW+AxdR5r17JvIsxbdX+nceEK
KWiS8YvkPJwlhIskMu8nBlc62efk6R8bVIRCrgbn87KNe/SmOTgUvgdw5zL5UxU7
m3IMdy7Cl9+0h7AYKUha2d05cAw5nEvmcJlOGjwygQKBgQD4vOuEJXcjjOYzOwEO
DbCfExCl9KnCCOJewq37AxBWreo3gWu4+S4RxSnsEN7NPQGj4JfePr/gSzcr0Zkb
wDZc1jVIUdh5eyE1ABvJWnyfYducKF1j5hO0XJNlHqg1+5DhtycsQRlsbiMDEUxk
1S/zMMg3Af/y87Su/wmnZdCo+QKBgQDTjzY2iaxhut3gi8vKzrS+YAAsjHw+ohT5
WVgFp+TP1lFEjV8hLhWWvnbfluFItzLawjYNpckNcHEA/cgTtsy2baEdrkhhFIj0
1FF2xIYJzHucHZT9e8hMU6FyoX/iqXSfA9bmc5LSV/Bi6nN8hneIcz/x/Vt1z3qd
EeUgHYnjWQKBgGwR2NnPVVYSz6mOh0TN2eEjbWZNSLxPE9tMBj8684xVf5+iEWWK
jeOWoEI6ijLtwJqs6A7dgIw44b2eEUGnX3cycm/7b2xIfQMECw6Oy/qLj9jnCLxw
qDsCxd93VGov5KDM7K4jkqIzr+6TQ3fD0FN+7F5J9iRekjA+Crm6WNAxAoGBAJkC
84rueCcXKHLHqVW9uywV8wpFcXc7c0AFRoyQqgVIVO7n8O3mjubASuncDoSxO67M
2Jt2VLvLn2/AHX1ksRsgn28AJolQeN3a0jC8YtWjd6OqIaBUbsIFmrd15zDgruBz
vnJfFMndoJdqSqy99KZT9OPpAsVqkpwX3UglFR3BAoGBAJLMwZ1bKqIH1BrZhSdx
dtDSoMoQsg+5mWVx5DXSyN4cgkykfbIqAPh8xe6hDFUwMBPniVj9D1c67YYPs/7/
9UtZHPN4s55Li7gJ4tGIpRkcThMEbdBE9rBzgFdNSPloBzwJgC4/XgWR6ZQr6zXD
CD/2ADbs1OybuNTkDSiPdw9K
-----END PRIVATE KEY-----
	`
	// Generated from
	// openssl rsa -in privkey.pem -pubout -out pubkey.pem
	// perl -pe 's/\n/\\n/g' pubkey.pem
	publicKeyPKCS8 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzY7cUlD+RY4g/WtClwcn
izxNcuXextVqSTqOmIACYxVBT2LPPhLRezM1qcb3EUBLmIK0eyuRO2ZkpLjN34Pp
FnrLKq3PKmS8dyxVb+wtjtiE9LNq4h4wWwToy2zND7Pbl+bub4MiDpRBs6NXcOQD
A4udkI430TjDC+1qkSzGO43EbiuyKS+VIZwLSpzq2pB0s8TB6iAtmUJ7CRLNkZ4+
4oYK+lWeuhRBFHCKm1MGQxB2LGppIOGwhRwfj/duRXmX0LQQwRpRRnjUrJgziq8O
uiNsA3avwWezDd8v7XaXWtB/veamEcnthXsi/bkqERFmgL+T3pkGBrdsHMZpDaCJ
kQIDAQAB
-----END PUBLIC KEY-----
`
)

var (
	testRequiredHeaders = []string{
		"X-Ops-Timestamp",
		"X-Ops-UserId",
		"X-Ops-Sign",
		"X-Ops-Content-Hash",
		"X-Ops-Authorization-1",
		"X-Ops-Request-Source",
	}

	mux      *http.ServeMux
	server   *httptest.Server
	client   *chef.Client
	service  *Service
	keyPairs = []keyPair{
		{
			privateKeyPKCS1,
			publicKeyPKCS1,
			"PKCS1",
		},
		{
			privateKeyPKCS8,
			publicKeyPKCS8,
			"PKCS8",
		},
	}
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	client, _ = chef.NewClient(&chef.Config{
		Name:                  userid,
		Key:                   privateKeyPKCS1,
		BaseURL:               server.URL,
		AuthenticationVersion: "1.0",
	})
	service = NewService(client)
}

func teardown() {
	server.Close()
}
