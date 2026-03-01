class PocketbaseMultiview < Formula
  desc "CLI tool for exploring multiple PocketBase instances"
  homepage "https://github.com/jiseop121/multi-pocketbase-ui"
  version "0.2.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-arm64.tar.gz"
      sha256 "e74731094bcd09b73ecffe813f8e097df97a443592bc9554d37b4001fc69a9e0"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.2.1/pbmulti-v0.2.1-darwin-amd64.tar.gz"
      sha256 "6f6214e6c630a1855e555fa818e1c76a6f0276ff15eb9684d6344333d99bd0de"
    end
  end

  def install
    bin.install "pbmulti"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/pbmulti -c \"version\"")
  end
end
