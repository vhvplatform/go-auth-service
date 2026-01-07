# Flutter - Mobile Application

This directory will contain the Flutter mobile application for the authentication service.

## Structure

```
flutter/
├── android/        # Android-specific code
├── ios/           # iOS-specific code
├── lib/           # Dart source code
│   ├── models/    # Data models
│   ├── screens/   # UI screens
│   ├── services/  # API services
│   ├── widgets/   # Reusable widgets
│   └── main.dart  # Entry point
├── test/          # Unit and widget tests
├── pubspec.yaml   # Dependencies
└── README.md      # This file
```

## Getting Started

(To be added when implementing the Flutter mobile app)

### Prerequisites

- Flutter SDK >= 3.0.0
- Dart SDK >= 3.0.0
- Android Studio (for Android development)
- Xcode (for iOS development on macOS)

### Installation

```bash
cd flutter
flutter pub get
```

### Development

```bash
flutter run
```

### Build

#### Android
```bash
flutter build apk
```

#### iOS
```bash
flutter build ios
```

## Features

- User authentication
- OAuth2 integration
- Biometric authentication support
- Secure token storage
- Role-based access control

## Technology Stack

- Flutter
- Dart
- Provider or Bloc for state management (TBD)
- Secure Storage for token management
