package config

// Version information for Pejelagarto Translator
// This constant is shared across all build modes.
//
// Build modes:
//   Backend server:  go build -tags "ngrok_default,downloadable" .
//   Frontend server: go build -tags "frontendserver,ngrok_default,downloadable" .
//   WASM module:     GOOS=js GOARCH=wasm go build -tags frontend .
//
// Changelog v1.2.7:
//   - FIXED: APK compatibility - now supports Android 4.1+ (API 16+) instead of 5.0+
//   - ENHANCED: Simplified AndroidManifest.xml for maximum device compatibility
//   - IMPROVED: APK includes all architectures (arm64-v8a, armeabi-v7a, x86, x86_64)
//   - REMOVED: Problematic manifest attributes causing installation failures
//   - VERIFIED: APK properly signed and compatible with wider range of devices
//
// Changelog v1.2.6:
//   - FIXED: APK signing with debug key for Android installation compatibility
//   - FIXED: AndroidManifest.xml (removed manual uses-sdk, let gomobile handle it)
//   - ENHANCED: APK now properly signed with v1, v2, and v3 signature schemes
//   - IMPROVED: Auto-creates debug keystore if not present for APK signing
//   - VERIFIED: APK installable on Android devices (API 21+ / Android 5.0+)
//
// Changelog v1.2.5:
//   - ADDED: Android APK download button (🤖 Android) to website
//   - AUTOMATED: Android APK build with gomobile (multi-arch: ARM64, ARMv7)
//   - AUTOMATED: Java JDK download (Microsoft OpenJDK 17) if not present
//   - AUTOMATED: Android SDK/NDK download and installation if not present
//   - ENHANCED: build-android-apk.ps1 now auto-downloads all prerequisites
//   - INTEGRATED: Android APK build into production build pipeline (step 3/6)
//   - REMOVED: Placeholder APK files - now builds real 13.57 MB functional APK
//
// Changelog v1.2.4:
//   - FIXED: Mobile download section layout (now appears below translator instead of to the right)
//   - ENHANCED: Body uses flex-direction: column for proper vertical stacking
//   - IMPROVED: Mobile layout with justify-content: flex-start and align-items: stretch
//   - SYNCED: Both main.go and server_frontend.go CSS are now consistent
//
// Changelog v1.2.3:
//   - FIXED: WASM loading issue on ngrok deployments ('Failed to load translation module')
//   - ENHANCED: server_frontend.go now searches multiple paths including executable directory
//   - IMPROVED: Build script copies WASM files to both bin/ and project root for reliability
//   - ADDED: path/filepath import for robust cross-platform path handling
//
// Changelog v1.2.2:
//   - FIXED: Version link CSS (better visibility in dark/light themes with badge style)
//   - FIXED: Pronunciation text box to always show Pejelagarto pronunciation
//   - FIXED: Removed incorrect number conversion in preprocessTextForTTS
//   - FIXED: Pronunciation updates correctly when swapping translation direction
//   - ADDED: RemoveTimestampSpecialCharacters exported function for TTS preprocessing
//
// Changelog v1.2.1:
//   - VALIDATED: All 13 build tag combinations compile and test successfully
//   - CONSOLIDATED: All garble+ngrok documentation into README.md
//   - CLEANUP: Removed separate MD files and log artifacts
//
// Changelog v1.2.0:
//   - FIXED: WASM file naming (translator.wasm)
//   - FIXED: ngrok domain configuration (empty for random URLs)
//   - DOCUMENTED: Garble+ngrok incompatibility (definitively proven)
//   - DEPRECATED: Garble production build for ngrok use
//
// Changelog v1.2.9:
//   - FIXED: Android WebView translation functionality (renamed translate() to doTranslate())
//   - FIXED: WebView JavaScript caching (added cache clearing on load)
//   - ENHANCED: Added WebChromeClient for better console logging and debugging
//   - IMPROVED: Android app now fully functional with bidirectional translation
//   - VERIFIED: JavaScript bridge successfully calls Go translation functions

// Version is the current version of the application
const Version = "1.2.9"
