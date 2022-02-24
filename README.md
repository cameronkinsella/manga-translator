# Manga Translator

An easy-to-use application for translating text in images in many languages.

The GUI was created using [Gio](https://gioui.org/). Gio supports a variety of platforms, including browsers, but this
application was designed to be used as a desktop application only.

> Note: This application should work on Windows, Linux, and macOS systems, but has only been tested on Windows.

<p align="center">
   <img src="https://github.com/cameronkinsella/manga-translator/blob/master/images/example-translation.gif?raw=true"  alt="Example Translation"/>
   <br/>
   <cite>Image source: うずまき 1巻</cite>
</p>

---

## Installation

### Option 1: Download the binary directly (Windows only)

Visit the [releases page](https://github.com/cameronkinsella/manga-translator/releases/latest/)
to download the latest version.

### Option 2: `go get`

```sh
# Manga Translator application
go get -u github.com/cameronkinsella/manga-translator/cmd/manga-translator

# Config setup application
go get -u github.com/cameronkinsella/manga-translator/cmd/manga-translator-setup
```

> Note: manga-translator creates a mtl-config.yml and mtl-cache.bin in the same directory as the binary,
> so consider changing your GOPATH to a more suitable directory before installing.

## Prerequisites

**Mandatory:**

- Google Cloud Vision API service account key

**At least one of the following are required:**

- Google Cloud Translation API key
- DeepL API key (Free or Pro)

---

### Google Cloud Vision API

Follow steps 1-5 of this guide :point_right:
[Cloud Vision API setup](https://cloud.google.com/vision/docs/before-you-begin).

This will create a service account key for the Vision API. The path to this JSON key will be needed to configure
manga-translator.

### Google Cloud Translation API

1. [Enable the Cloud Translation API](https://cloud.google.com/translate/docs/setup#api)
2. [Create an API key for the Cloud Translation API](https://cloud.google.com/docs/authentication/api-keys?hl=en#creating_an_api_key)

This API key will be needed to configure manga-translator (if you want to use this translation service).

### DeepL Translation API

1. Create an account on [deepl.com](https://deepl.com) and sign up for one of the
   [API plans](https://www.deepl.com/pro#developer)
2. Copy your API key from your account menu

This API key will be needed to configure manga-translator (if you want to use this translation service).

## Usage

### First Time (configuration)

Do one of the following:

1. Run the `manga-translator-setup` application and follow the interactive prompts
2. Create the `mtl-config.yml` file manually
   by [following the schema](./pkg/config/mtl-config.schema.yml)

### Command

```
Usage: manga-translator.exe [OPTIONS] [IMAGE_LOCATION]


Arguments:
  IMAGE_LOCATION   The path or URL of the image (not required if using -clip option).

Options:
  -url             Use an image from a URL instead of a local file.
  -clip            Use an image from your clipboard.
```

> Note: You can also open images with the `manga-translator` application itself
(on Windows, you can easily do this by dragging the image on top of `manga-translator.exe`)
