class Pbdash < Formula
  desc "Read-only CLI viewer for PocketBase instances"
  homepage "https://github.com/jiseop121/pbdash"
  version "0.5.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/pbdash/releases/download/v0.5.0/pbdash-v0.5.0-darwin-arm64.tar.gz"
      sha256 "abdc474359718ed69bfb15977307be2dbf280bbe574f7eb92e212e5a41a5cf3d"
    else
      url "https://github.com/jiseop121/pbdash/releases/download/v0.5.0/pbdash-v0.5.0-darwin-amd64.tar.gz"
      sha256 "b8e7bc2f6a269253f38207b2d0502b911ba58aa1eccf08da5c02a58e13b4c195"
    end
  end

  def install
    bin.install "pbdash"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbdash -c \"version\"")
  end
end
