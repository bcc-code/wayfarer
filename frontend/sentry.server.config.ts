import * as Sentry from "@sentry/nuxt";
 
Sentry.init({
  dsn: "https://a84ef55e626a228d04a301448f3e2973@o4507803294892032.ingest.de.sentry.io/4510569771368528",

  // We recommend adjusting this value in production, or using tracesSampler
  // for finer control
  tracesSampleRate: 1.0,

  // Enable logs to be sent to Sentry
  enableLogs: true,

  // Enable sending of user PII (Personally Identifiable Information)
  // https://docs.sentry.io/platforms/javascript/guides/nuxt/configuration/options/#sendDefaultPii
  sendDefaultPii: true,

  // Setting this option to true will print useful information to the console while you're setting up Sentry.
  debug: false,
});
