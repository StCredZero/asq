class Asq < Formula
  desc "Active Semantic Query Tool for Go code analysis"
  homepage "https://github.com/StCredZero/asq"
  version "0.1.0"
  url "https://github.com/StCredZero/asq/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "0" # TODO: Update this after first release
  head "https://github.com/StCredZero/asq.git", branch: "main"
  license "MIT"

  depends_on "go@1.23.4" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["GO111MODULE"] = "on"
    
    # Install dependencies
    system "go", "get", "github.com/go-enry/go-enry/v2"
    system "go", "get", "github.com/smacker/go-tree-sitter"
    system "go", "get", "github.com/alexflint/go-arg"
    
    system "go", "build", "-o", bin/"asq", "./cmd/asq"
  end

  test do
    (testpath/"test.go").write <<~EOS
      package main
      func main() {
        //asq_start
        println("Hello, World!")
        //asq_end
      }
    EOS
    
    assert_match "(call_expression", shell_output("#{bin}/asq tree-sitter #{testpath}/test.go")
    assert_match "//asq_match", shell_output("#{bin}/asq query #{testpath}/test.go")
  end
end
