# frozen_string_literal: true

# Copyright The OpenTelemetry Authors
#
# SPDX-License-Identifier: Apache-2.0

require 'rubygems'
require 'bundler/setup'
require 'action_controller/railtie'
require 'active_record'
require 'opentelemetry-instrumentation-rails'
require 'opentelemetry-instrumentation-active_record'
require 'uptrace'

# copy your project DSN here or use UPTRACE_DSN env var
Uptrace.configure_opentelemetry(dsn: '') do |c|
  c.use_all

  c.service_name = 'myservice'
  c.service_version = '1.0.0'
end

ActiveRecord::Base.establish_connection(
  adapter: 'sqlite3',
  database: 'db.sqlite3'
)
ActiveRecord::Base.logger = Logger.new(STDOUT)
ActiveRecord::Schema.define do
  create_table :posts, force: true do |t|
  end
end

class Post < ActiveRecord::Base
end

# TraceRequestApp is a minimal Rails application inspired by the Rails
# bug report template for action controller.
# The configuration is compatible with Rails 6.0
class TraceRequestApp < Rails::Application
  config.root = __dir__
  config.hosts << 'example.org'
  secrets.secret_key_base = 'secret_key_base'
  config.eager_load = false
  config.logger = Logger.new($stdout)
  Rails.logger  = config.logger

  routes.append do
    get '/', to: 'example#index'
    get '/hello/:username', to: 'example#hello', as: 'hello'
  end
end

# ExampleController
class ExampleController < ActionController::Base
  include Rails.application.routes.url_helpers

  def index
    Post.create

    trace_url = Uptrace.trace_url()
    render inline: %(
      <html>
        <p>Here are some routes for you:</p>
        <ul>
          <li><%= link_to 'Hello world', hello_path(username: 'world') %></li>
          <li><%= link_to 'Hello foo-bar', hello_path(username: 'foo-bar') %></li>
        </ul>
        <p>View trace: <a href="#{trace_url}" target="_blank">#{trace_url}</a></p>
      </html>
    )
  end

  def hello
    trace_url = Uptrace.trace_url()
    render inline: %(
      <html>
        <h3>Hello #{params[:username]}</h3>
        <p>View trace: <a href="#{trace_url}" target="_blank">#{trace_url}</a></p>
      </html>
    )
  end
end

Rails.application.initialize!

run Rails.application
