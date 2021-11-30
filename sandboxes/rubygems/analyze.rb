#!/usr/bin/env ruby
#
require 'find'
require 'open3'
require 'pathname'

def install(package, version, local_file)
  cmd = ["gem", "install"]
  if local_file
    cmd << local_file
  else
    if version
      cmd << "-v"
      cmd << version
    end
    cmd << package
  end

  output, status = Open3.capture2e(*cmd)
  puts output

  if status.success?
    puts "Install succeeded."
    return
  end

  puts "Install failed."
  if output.include? "Could not find a valid gem"
    exit 0
  end

  exit 1
end

if ARGV.length < 2 || ARGV.length > 4
  puts "Usage: #{$0} [--local file | --version version] phase package"
  exit 1
end

local_file = nil
version = nil

case ARGV[0]
when "--local"
  ARGV.shift
  local_file = ARGV.shift
when "--version"
  ARGV.shift
  version = ARGV.shift
end

phase = ARGV.shift
package = ARGV.shift

if phase != "all"
  puts "Only \"all\" phase is supported at the moment"
  exit 1
end

install(package, version, local_file)
spec = Gem::Specification.find_by_name(package)

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
