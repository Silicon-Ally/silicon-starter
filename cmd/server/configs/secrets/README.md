# Secrets

This directory stores configuration for the application that should be
kept secret, like database credentials, API keys for assorted services,
cryptographic key material, etc.

Files in here are JSON-formatted and encrypted with [sops](https://github.com/
mozilla/sops), see the [root `.sops.yaml` file](/.sops.yaml) for details.
