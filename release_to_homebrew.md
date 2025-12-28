---
description: How to release the app to Homebrew
---
1.  **Build Binary**:
    ```bash
    go build -o flowbuilder main.go
    tar -czf flowbuilder-v1.0.0.tar.gz flowbuilder
    ```
2.  **Calculate SHA256**:
    ```bash
    shasum -a 256 flowbuilder-v1.0.0.tar.gz
    ```
3.  **Create Github Release**: Upload the `.tar.gz`.
4.  **Create Formula**: `flowbuilder.rb`
    ```ruby
    class Flowbuilder < Formula
      desc "AI Multi-Agent Flowchart Builder"
      homepage "https://github.com/your-repo/flowbuilder"
      url "https://github.com/your-repo/flowbuilder/releases/download/v1.0.0/flowbuilder-v1.0.0.tar.gz"
      sha256 "REPLACE_WITH_SHA256"
      license "MIT"

      def install
        bin.install "flowbuilder"
      end

      test do
        system "#{bin}/flowbuilder", "--version"
      end
    end
    ```
5.  **Tap & Install**:
    ```bash
    brew tap your-username/homebrew-tap
    brew install flowbuilder
    ```
