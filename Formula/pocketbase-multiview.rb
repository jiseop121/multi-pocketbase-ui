class PocketbaseMultiview < Formula
  desc "CLI tool for exploring multiple PocketBase instances"
  homepage "https://github.com/jiseop121/multi-pocketbase-ui"
  version "0.2.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-arm64.tar.gz"
      sha256 "60d412ef660e931f7910a95a62e9422279bd279b0d8f9d2e58e1d213e43b19b9"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-amd64.tar.gz"
      sha256 "230cfd17739ade43085c969e6a8a66b478eebd994ac4649cf5c35e8e83dab3ec"
    end
  end

  def install
    bin.install "pbmulti"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbmulti -c \"version\"")
  end
end
