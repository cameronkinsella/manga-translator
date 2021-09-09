# Config Schema

If you wish to create `config.yml` manually, follow this schema:

```yaml
cloudVision:
  credentialsPath: string # Absolute path to service account key (json) for Cloud Vision
translation:
  selectedService: string # Selected translation service: 'deepL' or 'google'
  google:
    apiKey: string # Cloud Translation API key
  deepL:
    apiKey: string # DeepL API key
    premium: boolean # 'true' if the given API key is for a DeepL pro account, otherwise 'false'
```
