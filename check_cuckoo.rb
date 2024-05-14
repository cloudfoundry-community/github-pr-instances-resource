#!/usr/bin/env ruby

require 'octokit'
require 'json'

# this is a behavior switch to ease development and debugging
payload =
  JSON.parse(
    if File.exist? 'payload'
      File.open('payload', 'r') {|f| f.read }
    else
      ARGF.read
    end
  )

unless payload.has_key? 'source'
  STDERR.puts "Must pass 'source' on STDIN"
  exit 1
end

required_values = ['access_token', 'repository', 'number']
required_values.each do |value|
  unless payload['source'].has_key? value
    STDERR.puts "Must set source.#{value} on STDIN"
    exit 1
  end
end

c = Octokit::Client.new(access_token: payload['source']['access_token'], per_page: 100)
c.auto_paginate = true

commits = c.pull_request_commits(payload['source']['repository'], payload['source']['number'])

# kinda hate concourse for this.
# concourse can pass us a version or not, that may not exist remotely.
# 1. if there are new commits, we return the commit we were passed and all new ones. cool. this is sane.
# 2. concourse passed us a version, but we cant find it. fck, user rewrote the remote commit log deleting everything. we return everything we have, lol to concourse's immutability
# 3. this is the first run, concourse has no knowledge of versions, we pass everything we have.
new_commits =
  if payload['source'].has_key? 'version'
    if index = commits.find_index{|e| e[:sha] == payload['source']['version']['sha'] }
      commits[index..-1]        # 1
    else
      commits[-1..] || []       # 2
    end
  else
    commits[-1..] || []         # 3
  end

# This *was* set to match this https://github.com/telia-oss/github-pr-resource
# Now its set to match this: https://github.com/cloudfoundry-community/github-pr-instances-resource
new_commits_cleaned =
  new_commits.map do |e|
    {
      ref: e[:sha]
    }
  end

puts JSON.pretty_generate(new_commits_cleaned)

exit 0
