class PocketbaseMultiview < Formula
  desc "CLI tool for exploring multiple PocketBase instances"
  homepage "https://github.com/jiseop121/multi-pocketbase-ui"
  version "0.2.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-arm64.tar.gz"
      sha256 "01187c55efda1f189fa8d467405cc5c8bbb9a99a3966a0d486ece1af9059f400"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-amd64.tar.gz"
      sha256 "a5919c72a2047239e5d3a5594e2b15ea2085c2c0b6d81567315e84a38003db72"
    end
  end

  def install
    bin.install "pbmulti"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbmulti -c \"version\"")
  end
end
