./scripts/run_analysis.sh -ecosystem pypi -package pipx -mode dynamic -nopull > out.pypi.txt
./scripts/run_analysis.sh -ecosystem crates.io -package rand -mode dynamic -nopull > out.crates.txt
./scripts/run_analysis.sh -ecosystem rubygems -package ruby-macho -mode dynamic -nopull > out.ruby.txt
./scripts/run_analysis.sh -ecosystem npm -package chalk -mode dynamic -nopull > out.npm.txt