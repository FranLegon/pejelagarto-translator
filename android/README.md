# Android Bindings Package

This package provides Android bindings for the Pejelagarto translator library.

## Building

The Android APK is built using `gomobile` which cross-compiles Go code to Android native libraries.

### Prerequisites

1. Install gomobile:
```bash
go install golang.org/x/mobile/cmd/gomobile@latest
```

2. Initialize gomobile (downloads Android SDK/NDK):
```bash
gomobile init
```

### Build APK

To build the Android APK:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
export ANDROID_NDK_HOME=/usr/local/lib/android/sdk/ndk/26.3.11579264
export ANDROID_HOME=/usr/local/lib/android/sdk
gomobile build -androidapi=21 -target=android -o bin/pejelagarto-translator.apk ./cmd/androidapp
```

This creates a multi-architecture APK (~14MB) with native libraries for:
- ARMv7 (32-bit)
- ARM64 (64-bit)
- x86 (32-bit)
- x86_64 (64-bit)

## Package Structure

- `android/translator.go` - Go bindings that can be used with `gomobile bind` to create an AAR library
- `cmd/androidapp/` - Standalone Android app that can be built with `gomobile build`
  - `main.go` - Simple GL-based Android app that demonstrates the translator
  - `AndroidManifest.xml` - Android app manifest

## Usage

### As a Library (AAR)

Build an Android Archive (AAR) to use in Android Studio projects:

```bash
gomobile bind -androidapi=21 -target=android -o bin/pejelagarto-translator.aar ./android
```

Then import the AAR in your Android Studio project and use:

```java
import android.Android;

Android.Translator translator = Android.NewTranslator();
String translated = translator.TranslateToPejelagarto("Hello World");
```

### As a Standalone App (APK)

Install the APK on an Android device:

```bash
adb install bin/pejelagarto-translator.apk
```

The app provides a simple visual interface demonstrating the translation functionality.

## API

The translator provides these methods:

- `TranslateToPejelagarto(input string) string` - Translate from human to Pejelagarto
- `TranslateFromPejelagarto(input string) string` - Translate from Pejelagarto to human
- `ToPejelagarto(input string) string` - Convenience function for translation to Pejelagarto
- `FromPejelagarto(input string) string` - Convenience function for translation from Pejelagarto
