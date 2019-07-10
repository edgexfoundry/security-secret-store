
## Get a X.509 certificate file (PEM format)

```bash
root@vault-s1:/ # cat in.pem
-----BEGIN CERTIFICATE-----
MIIFrzCCA5egAwIBAgIFFRhkgFIwDQYJKoZIhvcNAQELBQAwgZAxCzAJBgNVBAYT
AkNIMQswCQYDVQQIDAJWRDERMA8GA1UEBwwITGF1c2FubmUxEzARBgNVBAoMClpl
bkVudHJvcHkxGzAZBgNVBAMMElplbkVudHJvcHkgVHJ1c3RDQTEvMC0GCSqGSIb3
DQEJARYgWmVuRW50cm9weVRydXN0Q0FAemVuZW50cm9weS5uZXQwHhcNMTgwMjE0
MjI0MDUzWhcNMjExMTEwMjI0MDUzWjCBiTELMAkGA1UEBhMCQ0gxCzAJBgNVBAgM
AlZEMREwDwYDVQQHDAhMYXVzYW5uZTETMBEGA1UECgwKWmVuRW50cm9weTEgMB4G
A1UEAwwXdmF1bHQtczEuemVuZW50cm9weS5uZXQxIzAhBgkqhkiG9w0BCQEWFGFk
bWluQHplbmVudHJvcHkubmV0MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKC
AgEArSTnMk8GWmE3merkT9tKuTToeUqAxMFlpVWNTi+abiFH87fSL/uas/2FVqLr
zSRxYioL7u0LiZpG4wtpbygl0zy0ayxT9oCtjOBZq2Rb/duvGtjf3tozC42IypbS
TJ+2lmxaGSpJEe7g5ntMgBT6Huk9OY68MBNBK01yUD2xDUvXRgCGn468ZlZZX+tW
YKHazYjvumzCXcPWQD5IEnInNpwaXBtFRwY37mz3GFOSmWMvj8AxT6GyDdVwOIvu
0Ttbh2oc/wq7/lQVcZgkU/NdZ2nELg7ACuzdGk3/ZE9vfA9u3DCGmaqTcpSNkWNa
m/LHtK2rMxGLAaTMxTAzmqUjr057/kugwrIayAta6MYmWT0HYbGLvM6hxB2erWkX
Jg5lnXac+q5GW9wJwsWeskGriIFOUKxZwvnbNFmTBke0wGVHxYDtUj4zA8yf8pSh
9vi5hF+mA+fcWATpDbS7f18PzBv7JJ1Ltq8QoKTMPEK+Pb/9OD9asAJ/N9xUVVcJ
rfBrvURXeUaAcwtS+H76ZLsswxMth/EDTBHlTbpFJC3OE1Lyxq9ST3fLQRxJFbMG
z0E9iHibADy1DZ1qdq+4PrK52s5P8aYoER0WdEoZB012hmgluf54P7QMU2dMzZBF
xBBlQTEfb7cVMrUttvvvejNlS0shK60B8bK4lHfWtTW89J0CAwEAAaMVMBMwEQYJ
YIZIAYb4QgEBBAQDAgZAMA0GCSqGSIb3DQEBCwUAA4ICAQBB4hY2UhL/2nXJQ+qL
hUzgvadCGa9kI/lv+8qZ34YZpKBuEO6ViCsRFqF/+CuHI+DNPTMqRuNCVYCjrgz4
l31PXDZqrQaJQvYtbwU2dmRQ7vc5mqchIajvbzLPoYJzDZ6mD4kBD+STxYisdAYd
gx/kTj50QhdQ77qYdgQ1lu1c3WWpgT/bHzj3s6N1zvAPxZUJJ11CJvQu/T0b1wiV
N9scPWoQvrrQVtvf2QDbV7J76gt81zpJN8+Dq66Flgw7hxFU7kw8A7t4UwOpQhKs
V8VwXFiSGeRz0lhJh4mOlVpoRhSuVOioYh1sF3xb/4dg2UJ1MF+0HDJzA+7DoR1C
Z7bPjdgG7vORJOQuFded5QLevn/gkL3oXPEa09xEFlcEBAGYgF9sRj6rI1SZ2J7u
M5ct06uRUp8wgxZ0blMbXse7P3pUMXS17SICEKteze8pmb193Q60d+r7XRZiYgOp
JPOmtL34fPVeI4/CIzW/rivBUDd9fUIJfuMaBpA5+eMfHrGBBmXOoTJORsBB48+Y
6XKWd1B1uzoP7pOIZMBFBDqvBbLJhs4dQnbW38stzcVqRddbNV7K/CkysRF6uAGM
/jSh0mLIfaf6Yz6xqz19t7sssgl549qkVL8aTwZD3wF7RyUNISMuKa/VQl03z1rq
/Q96QRSCeQGi35SGbCAvIMwWzw==
-----END CERTIFICATE-----
```

## Produce a hash fingerprint using SHA-256 algo
> We will need this for an easy comparaison method

```bash
root@vault-s1:/ # sha256sum in.pem
76215d82a486a38593a117f2d53d17014333234db34fb1406e4187fe6ed8db61  in.pem
```

## Be extremely carefull the PEM file does not to contain \n or \r as last byte

```bash
root@vault-s1:/ # xxd in.pem | tail
00000760: 7a63 5671 5264 6462 4e56 374b 2f43 6b79  zcVqRddbNV7K/Cky
00000770: 7352 4636 7541 474d 0a2f 6a53 6830 6d4c  sRF6uAGM./jSh0mL
00000780: 4966 6166 3659 7a36 7871 7a31 3974 3773  Ifaf6Yz6xqz19t7s
00000790: 7373 676c 3534 3971 6b56 4c38 6154 775a  ssgl549qkVL8aTwZ
000007a0: 4433 7746 3752 7955 4e49 534d 754b 612f  D3wF7RyUNISMuKa/
000007b0: 5651 6c30 337a 3172 710a 2f51 3936 5152  VQl03z1rq./Q96QR
000007c0: 5343 6551 4769 3335 5347 6243 4176 494d  SCeQGi35SGbCAvIM
000007d0: 7757 7a77 3d3d 0a2d 2d2d 2d2d 454e 4420  wWzw==.-----END
000007e0: 4345 5254 4946 4943 4154 452d 2d2d 2d2d  CERTIFICATE-----
000007f0: 0a                                       .
```

## Write the certificate file (PEM) in Vault
> Vault secret root: "secret" (default)
> Vault certificate root: "secret/cert" (manually created)
> (k.v) = (cert1, PEM file)
>
> 2 possible syntaxes:

```bash
root@vault-s1:/ # cat in.pem | vault write secret/cert/cert1 value=-
Success! Data written to: secret/cert/cert1

# or

root@vault-s1:/ # vault write secret/cert/cert1 value=@in.pem
Success! Data written to: secret/cert/cert1
```

## List the keys under Vault secret/cert path

```bash
root@vault-s1:/ # vault list secret/cert
Keys
----
cert1
```

## Read the key "cert1" and get the corresponding PEM file
> Simply redirect the vault cmd output

```bash
root@vault-s1:/ # vault read -field=value secret/cert/cert1 > out.pem
```

## Double check if the original certificate file is equally identical to the one read

```bash
root@vault-s1:/ # diff in.pem out.pem
root@vault-s1:/ #
root@vault-s1:/ # cmp in.pem out.pem
root@vault-s1:/ # 

root@vault-s1:/ # sha256sum out.pem
76215d82a486a38593a117f2d53d17014333234db34fb1406e4187fe6ed8db61  out.pem
```
