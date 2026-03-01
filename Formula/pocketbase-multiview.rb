class PocketbaseMultiview < Formula
  desc "CLI tool for exploring multiple PocketBase instances"
  homepage "https://github.com/jiseop121/multi-pocketbase-ui"
  version "0.2.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-arm64.tar.gz"
      sha256 "b8b50e7b184eeb401128c91c6eca548201407d7b228de218b23c6e3ab95ea92b"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-amd64.tar.gz"
      sha256 "79b2b8bd6015c0b3e19022a546e980ed1db84bba4f53afbee57ceb5f929a1d37"
    end
  end

  def install
    bin.install "pbmulti"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbmulti -c \"version\"")
  end
end
