// Package logback owns the logback-spring.xml the connector reads, and the
// in-container path it is mounted at.
//
// It is its own package because all three renderers need it: kubernetes puts
// the content in a ConfigMap key, docker inlines it as a compose config, and
// podman bind-mounts it from a file. Having dockergen or podmangen reach into
// internal/deploy for it would be a cross-module import between sibling
// renderers, so the payload and its path live here instead -- the same shape
// internal/statusscript already uses for the status script.
package logback

import "github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"

// ContainerPath is where the connector expects the file, on every platform. It
// is spelled once here so moving it is one edit rather than three.
const ContainerPath = "/app/external/classpath/logback-spring.xml"

// FileName is the last element of ContainerPath, for the platforms that write
// the file to disk before mounting it.
const FileName = "logback-spring.xml"

// XML returns the logback-spring.xml content for the given syslog
// protocol. Host/port/appname flow in at runtime via springProperty from the
// LOGGING_SYSLOG_* env vars set on the Deployment, so the XML itself is static.
func XML(protocol string) string {
	if protocol == spec.SyslogTCP {
		return logbackTCP
	}
	return logbackUDP
}

// logbackUDP mirrors the reference deployment: Logback's built-in UDP
// (RFC 3164) SyslogAppender behind an async wrapper.
const logbackUDP = `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <!-- Spring Boot defaults: conversion rules, CONSOLE appender -->
  <include resource="org/springframework/boot/logging/logback/defaults.xml"/>
  <include resource="org/springframework/boot/logging/logback/console-appender.xml"/>

  <!-- Spring property substitution from application.yml or env vars -->
  <springProperty scope="context" name="SYSLOG_HOST" source="logging.syslog.host" />
  <springProperty scope="context" name="SYSLOG_PORT" source="logging.syslog.port" />
  <springProperty scope="context" name="SYSLOG_APPNAME" source="logging.syslog.appname" defaultValue="solace-ibmmq-connector-default"/>

  <!-- UDP syslog, RFC 3164. This is Logback's built-in, no extra deps. -->
  <appender name="SYSLOG" class="ch.qos.logback.classic.net.SyslogAppender">
    <syslogHost>${SYSLOG_HOST}</syslogHost>
    <port>${SYSLOG_PORT}</port>
    <facility>LOCAL0</facility>
    <suffixPattern>${SYSLOG_APPNAME}[%thread] %-5level %logger{36} - %msg</suffixPattern>
    <stackTracePattern>${SYSLOG_APPNAME}[%thread] \t</stackTracePattern>
    <throwableExcluded>false</throwableExcluded>
  </appender>

  <!-- Async wrapper. With UDP this is less critical (sends never block long)
   but still good practice - keeps any DNS resolution or socket setup off
   the app threads. -->
  <appender name="ASYNC_SYSLOG" class="ch.qos.logback.classic.AsyncAppender">
    <appender-ref ref="SYSLOG"/>
    <queueSize>2048</queueSize>
    <discardingThreshold>0</discardingThreshold>
    <neverBlock>true</neverBlock>
    <includeCallerData>false</includeCallerData>
  </appender>

  <root level="INFO">
    <appender-ref ref="CONSOLE"/>
    <appender-ref ref="ASYNC_SYSLOG"/>
  </root>
</configuration>
`

// logbackTCP sends newline-framed lines over TCP via logstash-logback-encoder's
// LogstashTcpSocketAppender (internally async, so no AsyncAppender wrapper).
// The jar must be provided on the connector classpath, e.g. via libs.
const logbackTCP = `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <include resource="org/springframework/boot/logging/logback/defaults.xml"/>
  <include resource="org/springframework/boot/logging/logback/console-appender.xml"/>

  <springProperty scope="context" name="SYSLOG_HOST" source="logging.syslog.host" />
  <springProperty scope="context" name="SYSLOG_PORT" source="logging.syslog.port" />
  <springProperty scope="context" name="SYSLOG_APPNAME" source="logging.syslog.appname" defaultValue="solace-ibmmq-connector-default"/>

  <!-- TCP syslog via logstash-logback-encoder (jar required on the classpath). -->
  <appender name="SYSLOG" class="net.logstash.logback.appender.LogstashTcpSocketAppender">
    <destination>${SYSLOG_HOST}:${SYSLOG_PORT}</destination>
    <encoder class="ch.qos.logback.classic.encoder.PatternLayoutEncoder">
      <pattern>${SYSLOG_APPNAME}[%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>

  <root level="INFO">
    <appender-ref ref="CONSOLE"/>
    <appender-ref ref="SYSLOG"/>
  </root>
</configuration>
`
