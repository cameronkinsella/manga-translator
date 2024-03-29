# Manga Translator

An easy-to-use application for translating text in images in many languages.

The GUI was created using [Gio](https://gioui.org/). Gio supports a variety of platforms, including browsers, but this
application was designed to be used as a desktop application only.
<p align="center">
   <img src="./images/example-translation.gif"  alt="Example Translation"/>
   <br/>
   <cite>Image source: うずまき 1巻</cite>
</p>

---

## Installation

### Option 1: Download the binary directly (Windows only)

Visit the [releases page](https://github.com/cameronkinsella/manga-translator/releases/latest/)
to download the latest version.

### Option 2: `go install`

**Manga Translator application:**

```sh
go install github.com/cameronkinsella/manga-translator/cmd/manga-translator@latest
```

**Config setup application:**

```sh
go install github.com/cameronkinsella/manga-translator/cmd/manga-translator-setup@latest
```

## Prerequisites

**Mandatory:**

- Google Cloud Vision API service account key

**At least one of the following are required:**

- Google Cloud Translation API enabled
- DeepL API key (Free or Pro)

---

### [Google Cloud Vision API](https://cloud.google.com/vision/docs/before-you-begin)

Quick guide:

1. [Enable the Cloud Vision API](https://console.cloud.google.com/flows/enableapi?apiid=vision.googleapis.com)
2. [Enable the IAM API](https://console.cloud.google.com/flows/enableapi?apiid=iam.googleapis.com)
3. Press the button below to create a new service account key:

   [![Open in Cloud Shell][shell_img]][sa_key]

This will create a service account key for the Vision API. The path to this JSON key will be needed to configure
manga-translator.

### [Google Cloud Translation API](https://cloud.google.com/translate/docs/setup)

Quick guide:

- [Enable the Cloud Translation API](https://console.cloud.google.com/flows/enableapi?apiid=translate.googleapis.com)

If you are using the same project as your Cloud Vision API and service key, then that is all.

If you are using a different project for the Cloud Vision API, you must also do the following:

- Press the button below to create a
  new [Cloud Vision API key](https://cloud.google.com/docs/authentication/api-keys?hl=en#creating_an_api_key):

  [![Open in Cloud Shell][shell_img]][api_key]

### DeepL Translation API

1. Create an account on [deepl.com](https://deepl.com) and sign up for one of the
   [API plans](https://www.deepl.com/pro#developer)
2. Copy your API key from your account menu

This API key will be needed to configure manga-translator (if you want to use this translation service).

## Usage

### First Time (configuration)

Do one of the following:

1. Run the `manga-translator-setup` application and follow the interactive prompts
2. Create the `mtl/mtl-config.yml` file manually by [following the schema](./pkg/config/mtl-config.schema.yml)

### Command

```
Usage: manga-translator.exe [OPTIONS] [IMAGE_LOCATION] [IMAGE_LOCATION]...


Arguments:
  IMAGE_LOCATION   The path or URL of the image (not required if using -clip option).

Options:
  -url             Use images from URLs instead of local files.
  -clip            Use an image from your clipboard.
```

> Note: On Windows you can also open it by dragging images on top of `manga-translator.exe`

### GUI

Coloured boxes will appear around all the text that was detected. Click on those boxes to display the original text and
the translation of that text.

You can click on the text in the "Original Text" or "Translated Text" sections to copy that text to your clipboard.

If you selected multiple images, you can navigate through them using the left and right arrow keys or the A and D keys.

[shell_img]: https://gstatic.com/cloudssh/images/open-btn.png

[sa_key]: https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/cameronkinsella/manga-translator&open_in_editor=dist/cloudshell/create-service-account-key.md

[api_key]: https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/cameronkinsella/manga-translator&open_in_editor=dist/cloudshell/create-translation-api-key.md
