import './components/ChainSearch.js';
import './components/TimeSeriesChart.js';
import './components/HeatmapChart.js';
import './components/TimeRangeSelector.js';
import './components/LogInBtn.js';
import './components/SubscriptionModal.js';
import './components/ThemeToggle.js';
import './lib/auth.js';

// Initialize theme from localStorage
const savedTheme = localStorage.getItem('theme') || 'light';
document.documentElement.setAttribute('data-theme', savedTheme);
document.querySelector('meta[name="color-scheme"]').setAttribute('content', savedTheme);

// Update Pico CSS theme
const link = document.querySelector('link[href*="pico"]');
const newHref = link.href.replace(/pico\.[^.]+\.css/, `pico.${savedTheme}.min.css`);
link.href = newHref;

