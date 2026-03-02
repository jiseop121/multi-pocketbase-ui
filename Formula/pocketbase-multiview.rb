class PocketbaseMultiview < Formula
  desc "CLI tool for exploring multiple PocketBase instances"
  homepage "https://github.com/jiseop121/multi-pocketbase-ui"
  version "0.3.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.3.1/pbviewer-v0.3.1-darwin-arm64.tar.gz"
      sha256 "202f9b3b4697d4b8bbcfbbdd3ca078c92c0012574591cfa84f9056de7cf76889"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.3.1/pbviewer-v0.3.1-darwin-amd64.tar.gz"
      sha256 "561fb4d5359d428da8972ff2fcd52c0f5ccd38a26abb4e11d3d2663b3aebac70"
    end
  end

  def install
    bin.install "pbviewer"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbviewer -c \"version\"")
  end
end
