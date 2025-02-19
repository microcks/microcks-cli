class Microcks < Formula
    desc "Simple CLI for interacting with Microcks test APIs"
    homepage "https://github.com/microcks/microcks-cli"
    url "https://github.com/microcks/microcks-cli/archive/refs/tags/0.5.7.tar.gz"
    sha256 "be49890c386736def0ba2ce1ddc002d49efaad04cc06188cd7f27593096274ea"
    license "Apache-2.0"
  
    depends_on "go" => :build
  
    def install
      ENV["CGO_ENABLED"] = "0"

      system "go", "build", "./main.go"    
    end
  
    test do 
        
    end
  
  end