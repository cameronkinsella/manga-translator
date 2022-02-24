# Config Schema

If you wish to create `mtl-config.yml` manually, follow the [config schema](./mtl-config.schema.yml).

Example config:

```yaml
# Example config
cloudVision:
  credentialsPath: C:\Users\me\credentials.json # Absolute path to service account key (json) for Cloud Vision
translation:
  selectedService: deepL # Selected translation service: 'deepL' or 'google'
  google:
    apiKey: abcdef123456 # Cloud Translation API key
  deepL:
    apiKey: abcdef123456 # DeepL API key
    premium: true # 'true' if the given API key is for a DeepL pro account, otherwise 'false'
```
