# Config Schema

If you wish to create `mtl/mtl-config.yml` manually, follow the [config schema](./mtl-config.schema.yml).

The `mtl` directory which contains this config file should be in the same directory as the main application.

Example config:

```yaml
# Example config
cloudVision:
  credentialsPath: C:\Users\me\credentials.json # Absolute path to service account key (json) for Cloud Vision
translation:
  selectedService: deepL # Selected translation service: 'deepL' or 'google'
  sourceLanguage: JA # OPTIONAL: The source language ISO-639-1 code. If omitted, the source language is automatically detected.
  targetLanguage: EN-US # The target language ISO-639-1 code.
  google:
    apiKey: abcdef123456 # Cloud Translation API key
  deepL:
    apiKey: abcdef123456 # DeepL API key
```
