class Pbdash < Formula
  desc "Read-only CLI viewer for PocketBase instances"
  homepage "https://github.com/jiseop121/pbdash"
  version "0.4.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/pbdash/releases/download/v0.4.0/pbdash-v0.4.0-darwin-arm64.tar.gz"
      sha256 "REPLACE_ME_WITH_RELEASE_SHA_ARM64"
    else
      url "https://github.com/jiseop121/pbdash/releases/download/v0.4.0/pbdash-v0.4.0-darwin-amd64.tar.gz"
      sha256 "REPLACE_ME_WITH_RELEASE_SHA_AMD64"
    end
  end

  def install
    bin.install "pbdash"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbdash -c \"version\"")
  end
end
