# prometheus-imap-exporter

Prometheus monitoring of mailboxes through IMAP

## Install

Having a working Golang environment:

```bash
go install github.com/trazfr/prometheus-imap-exporter@latest
```

## Example of configuration files

- `timeout` is in seconds
- `imap.cacert.example.com` uses a CACert certificate which is not installed on our computer, so we provide the [class 3 CAcert.org certificate](http://www.cacert.org/index.php?id=3)
- `imap.free.fr` uses the system's TLS certificates

```
{
    "listen": ":9091",
    "timeout": 5.0,
    "accounts": [
        {
            "url": "imaps://user:password@imap.cacert.example.com",
            "pem": "-----BEGIN CERTIFICATE-----\nMIIG0jCCBLqgAwIBAgIBDjANBgkqhkiG9w0BAQsFADB5MRAwDgYDVQQKEwdSb290\nIENBMR4wHAYDVQQLExVodHRwOi8vd3d3LmNhY2VydC5vcmcxIjAgBgNVBAMTGUNB\nIENlcnQgU2lnbmluZyBBdXRob3JpdHkxITAfBgkqhkiG9w0BCQEWEnN1cHBvcnRA\nY2FjZXJ0Lm9yZzAeFw0xMTA1MjMxNzQ4MDJaFw0yMTA1MjAxNzQ4MDJaMFQxFDAS\nBgNVBAoTC0NBY2VydCBJbmMuMR4wHAYDVQQLExVodHRwOi8vd3d3LkNBY2VydC5v\ncmcxHDAaBgNVBAMTE0NBY2VydCBDbGFzcyAzIFJvb3QwggIiMA0GCSqGSIb3DQEB\nAQUAA4ICDwAwggIKAoICAQCrSTURSHzSJn5TlM9Dqd0o10Iqi/OHeBlYfA+e2ol9\n4fvrcpANdKGWZKufoCSZc9riVXbHF3v1BKxGuMO+f2SNEGwk82GcwPKQ+lHm9WkB\nY8MPVuJKQs/iRIwlKKjFeQl9RrmK8+nzNCkIReQcn8uUBByBqBSzmGXEQ+xOgo0J\n0b2qW42S0OzekMV/CsLj6+YxWl50PpczWejDAz1gM7/30W9HxM3uYoNSbi4ImqTZ\nFRiRpoWSR7CuSOtttyHshRpocjWr//AQXcD0lKdq1TuSfkyQBX6TwSyLpI5idBVx\nbgtxA+qvFTia1NIFcm+M+SvrWnIl+TlG43IbPgTDZCciECqKT1inA62+tC4T7V2q\nSNfVfdQqe1z6RgRQ5MwOQluM7dvyz/yWk+DbETZUYjQ4jwxgmzuXVjit89Jbi6Bb\n6k6WuHzX1aCGcEDTkSm3ojyt9Yy7zxqSiuQ0e8DYbF/pCsLDpyCaWt8sXVJcukfV\nm+8kKHA4IC/VfynAskEDaJLM4JzMl0tF7zoQCqtwOpiVcK01seqFK6QcgCExqa5g\neoAmSAC4AcCTY1UikTxW56/bOiXzjzFU6iaLgVn5odFTEcV7nQP2dBHgbbEsPyyG\nkZlxmqZ3izRg0RS0LKydr4wQ05/EavhvE/xzWfdmQnQeiuP43NJvmJzLR5iVQAX7\n6QIDAQABo4IBiDCCAYQwHQYDVR0OBBYEFHWocWBMiBPweNmJd7VtxYnfvLF6MA8G\nA1UdEwEB/wQFMAMBAf8wXQYIKwYBBQUHAQEEUTBPMCMGCCsGAQUFBzABhhdodHRw\nOi8vb2NzcC5DQWNlcnQub3JnLzAoBggrBgEFBQcwAoYcaHR0cDovL3d3dy5DQWNl\ncnQub3JnL2NhLmNydDBKBgNVHSAEQzBBMD8GCCsGAQQBgZBKMDMwMQYIKwYBBQUH\nAgEWJWh0dHA6Ly93d3cuQ0FjZXJ0Lm9yZy9pbmRleC5waHA/aWQ9MTAwNAYJYIZI\nAYb4QgEIBCcWJWh0dHA6Ly93d3cuQ0FjZXJ0Lm9yZy9pbmRleC5waHA/aWQ9MTAw\nUAYJYIZIAYb4QgENBEMWQVRvIGdldCB5b3VyIG93biBjZXJ0aWZpY2F0ZSBmb3Ig\nRlJFRSwgZ28gdG8gaHR0cDovL3d3dy5DQWNlcnQub3JnMB8GA1UdIwQYMBaAFBa1\nMhvUx/Pg5o7zvdKwOu6yORjRMA0GCSqGSIb3DQEBCwUAA4ICAQBakBbQNiNWZJWJ\nvI+spCDJJoqp81TkQBg/SstDxpt2CebKVKeMlAuSaNZZuxeXe2nqrdRM4SlbKBWP\n3Rn0lVknlxjbjwm5fXh6yLBCVrXq616xJtCXE74FHIbhNAUVsQa92jzQE2OEbTWU\n0D6Zghih+j+cN0eFiuDuc3iC1GuZMb/Zw21AXbkVxzZ4ipaL0YQgsSt1P22ipb69\n6OLkrURctgY2cHS4pI62VpRgkwJ/Lw2n+C9vtukozMhrlPSTA0OhNEGiGp2hRpWa\nhiG+HGcIYfAV9v7og3dO9TnS0XDbbk1RqXPpc/DtrJWzmZN0O4KIx0OtLJJWG9zp\n9JrJyO6USIFYgar0U8HHHoTccth+8vJirz7Aw4DlCujo27OoIksg3OzgX/DkvWYl\n0J8EMlXoH0iTv3qcroQItOUFsgilbjRba86Q5kLhnCxjdW2CbbNSp8vlZn0uFxd8\nspxQcXs0CIn19uvcQIo4Z4uQ+00Lg9xI9YFV9S2MbSanlNUlvbB4UvHkel0p6bGt\nAmp1dJBSkZOFm0Z6ek+G7w7R1aTifjGJrdw032O+VIKwCgu8DdskR0w0B68ydZn0\nATnMnr5ExvcWkZBtCgQa2NvSKrcQnlaqo9icEF4XevI/VTezlb1LjYMWHVd5R6C2\np4wTyVBIM8hjrLcKiChF43GRJtne7w==\n-----END CERTIFICATE-----\n"
        },
        {
            "url": "imaps://user2:password2@imap.free.fr"
        }
    ]
}
```
