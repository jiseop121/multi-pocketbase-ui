class Pbdash < Formula
  desc "Read-only CLI viewer for PocketBase instances"
  homepage "https://github.com/jiseop121/pbdash"
  version "0.5.2"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/pbdash/releases/download/v0.5.2/pbdash-v0.5.2-darwin-arm64.tar.gz"
      sha256 "f3bf7b1d469d62e2bf986835ec130bdded0e3d73f8de61e45cb6f105a1726e86"
    else
      url "https://github.com/jiseop121/pbdash/releases/download/v0.5.2/pbdash-v0.5.2-darwin-amd64.tar.gz"
      sha256 "95603db29aa2498beb3f00ff45746fe6df2fa2dbc8ace8525054cb5e475b72f5"
    end
  end

  def install
    bin.install "pbdash"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbdash -c \"version\"")
  end
end
