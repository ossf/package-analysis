#!/usr/bin/env ruby
#
require 'find'
require 'open3'
require 'pathname'

class Package
  attr_reader :name, :version, :local_file

  def initialize(name:, version:, local_file:)
    @name, @version, @local_file = name, version, local_file
  end
end

def install(package)
  cmd = ["gem", "install"]
  if package.local_file
    cmd << package.local_file
  else
    if package.version
      cmd << "-v"
      cmd << package.version
    end
    cmd << package.name
  end

  output, status = Open3.capture2e(*cmd)
  puts output

  if status.success?
    puts "Install succeeded."
    return
  end

  # Always exit on failure.
  # Install failing is either an interesting issue, or an opportunity to
  # improve the analysis.
  puts "Install failed."
  exit 1
end

def importPkg(package)
  spec = Gem::Specification.find_by_name(package.name)

  spec.require_paths.each do |require_path|
    if Pathname.new(require_path).absolute?
      lib_path = Pathname.new(require_path)
    else
      lib_path = Pathname.new(File.join(spec.full_gem_path, require_path))
    end

    Find.find(lib_path.to_s) do |path|
      if path.end_with?('.rb')
        relative_path = Pathname.new(path).relative_path_from(lib_path)

        require_path = relative_path.to_s.delete_suffix('.rb')
        puts "Loading #{require_path}"
        begin
          require require_path
        rescue Exception => e
          puts "Failed to load #{require_path}: #{e}"
        end
      end
    end
  end
end

phases = {
  "all" => [method(:install), method(:importPkg)],
  "install" => [method(:install)],
  "import" => [method(:importPkg)],
}

if ARGV.length < 2 || ARGV.length > 4
  puts "Usage: #{$0} [--local file | --version version] phase package"
  exit 1
end

local_file = nil
version = nil

# Parse the arguments manually to avoid introducing unnecessary dependencies
# and side effects that add noise to the strace output.
case ARGV[0]
when "--local"
  ARGV.shift
  local_file = ARGV.shift
when "--version"
  ARGV.shift
  version = ARGV.shift
end

phase = ARGV.shift
package_name = ARGV.shift

package = Package.new(name: package_name, version: version, local_file: local_file)

if !phases.has_key?(phase)
  puts "Unknown phase #{phase} specified"
  exit 1
end

phases[phase].each { |m| m.call(package) }
