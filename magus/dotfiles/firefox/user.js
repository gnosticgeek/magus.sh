// Firefox user.js — conservative privacy + de-bloat. Restart Firefox to apply.

// Telemetry & data collection.
user_pref("toolkit.telemetry.enabled", false);
user_pref("toolkit.telemetry.unified", false);
user_pref("toolkit.telemetry.archive.enabled", false);
user_pref("datareporting.healthreport.uploadEnabled", false);
user_pref("datareporting.policy.dataSubmissionEnabled", false);
user_pref("app.shield.optoutstudies.enabled", false);
user_pref("app.normandy.enabled", false);

// Crash reports.
user_pref("browser.tabs.crashReporting.sendReport", false);
user_pref("browser.crashReports.unsubmittedCheck.autoSubmit2", false);

// New-tab de-bloat: no sponsored tiles, stories, or telemetry.
user_pref("browser.newtabpage.activity-stream.showSponsored", false);
user_pref("browser.newtabpage.activity-stream.showSponsoredTopSites", false);
user_pref("browser.newtabpage.activity-stream.feeds.section.topstories", false);
user_pref("browser.newtabpage.activity-stream.telemetry", false);

// Pocket & feature recommendations.
user_pref("browser.discovery.enabled", false);
user_pref("extensions.htmlaboutaddons.recommendations.enabled", false);

// Address bar: no sponsored / quick-suggest results.
user_pref("browser.urlbar.suggest.quicksuggest.sponsored", false);
user_pref("browser.urlbar.suggest.quicksuggest.nonsponsored", false);
user_pref("browser.urlbar.quicksuggest.enabled", false);

// Mild privacy — safe, no site breakage.
user_pref("privacy.trackingprotection.enabled", true);
user_pref("privacy.globalprivacycontrol.enabled", true);

// Quality of life.
user_pref("browser.startup.page", 3);
