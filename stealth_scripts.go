package goddgs

// NOTE: these scripts are defensive browser-normalization shims. They avoid
// unstable automation flags and keep page runtime behavior closer to regular
// browsing environments.

const scriptHideWebdriver = `(() => {
  const d = Object.getOwnPropertyDescriptor(Navigator.prototype, 'webdriver');
  if (!d || d.configurable) {
    Object.defineProperty(Navigator.prototype, 'webdriver', {
      get: () => undefined,
      configurable: true,
    });
  }
})();`

const scriptLanguageAndPlatform = `(() => {
  const langs = ['en-US', 'en'];
  Object.defineProperty(Navigator.prototype, 'languages', {
    get: () => langs,
    configurable: true,
  });
  Object.defineProperty(Navigator.prototype, 'platform', {
    get: () => 'Win32',
    configurable: true,
  });
})();`

const scriptChromeRuntime = `(() => {
  if (!window.chrome) {
    Object.defineProperty(window, 'chrome', {
      value: { runtime: {} },
      configurable: true,
    });
  }
})();`

const scriptPermissionsConsistency = `(() => {
  const originalQuery = window.navigator.permissions && window.navigator.permissions.query;
  if (!originalQuery) return;
  window.navigator.permissions.query = (parameters) => {
    if (parameters && parameters.name === 'notifications') {
      return Promise.resolve({ state: Notification.permission });
    }
    return originalQuery(parameters);
  };
})();`

const scriptHardwareHints = `(() => {
  Object.defineProperty(Navigator.prototype, 'hardwareConcurrency', {
    get: () => 8,
    configurable: true,
  });
  Object.defineProperty(Navigator.prototype, 'deviceMemory', {
    get: () => 8,
    configurable: true,
  });
})();`

// StealthScripts returns the JS snippets for a selected level.
func StealthScripts(level StealthLevel) []string {
	scripts := []string{scriptHideWebdriver, scriptLanguageAndPlatform}
	switch level {
	case StealthLevelAggressive:
		scripts = append(scripts, scriptChromeRuntime, scriptPermissionsConsistency, scriptHardwareHints)
	case StealthLevelStrong:
		scripts = append(scripts, scriptChromeRuntime, scriptPermissionsConsistency)
	case StealthLevelBasic:
		// basic only keeps minimal normalisation.
	default:
		scripts = append(scripts, scriptChromeRuntime, scriptPermissionsConsistency)
	}
	out := make([]string, len(scripts))
	copy(out, scripts)
	return out
}
