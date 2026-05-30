class OcGoCcPlus < Formula
  desc "Proxy Claude Code verso OpenCode Go con preset, sync modelli e routing endpoint"
  homepage "https://github.com/senseoverflow/oc-go-cc-plus"
  version "0.2.0"
  license "AGPL-3.0-only"

  on_macos do
    on_arm do
      url "https://github.com/senseoverflow/oc-go-cc-plus/releases/download/v0.2.0/oc-go-cc-plus_darwin-arm64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
    on_intel do
      url "https://github.com/senseoverflow/oc-go-cc-plus/releases/download/v0.2.0/oc-go-cc-plus_darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/senseoverflow/oc-go-cc-plus/releases/download/v0.2.0/oc-go-cc-plus_linux-arm64"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/senseoverflow/oc-go-cc-plus/releases/download/v0.2.0/oc-go-cc-plus_linux-amd64"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "oc-go-cc-plus"
  end

  def caveats
    <<~EOS
      Configurazione: ~/.config/oc-go-cc-plus/config.json

      Setup rapido:
        export OC_GO_CC_PLUS_API_KEY="your-opencode-go-key"
        oc-go-cc-plus init --preset deepseek
        oc-go-cc-plus doctor
        oc-go-cc-plus serve

      Claude Code:
        export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
        export ANTHROPIC_AUTH_TOKEN=unused
    EOS
  end

  test do
    assert_match "oc-go-cc-plus", shell_output("#{bin}/oc-go-cc-plus --version")
    assert_match "deepseek", shell_output("#{bin}/oc-go-cc-plus preset list")
  end
end
