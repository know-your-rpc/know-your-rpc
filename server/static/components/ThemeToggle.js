//@ts-nocheck
class ThemeToggle extends HTMLElement {
    constructor() {
        super();
        this.container = document.createElement("container");
        this.container.innerHTML = `
            <a id="theme-toggle">
                theme:🌙
            </a>
        `;
        this.appendChild(this.container);
        this.btn = document.getElementById("theme-toggle");
        this.attachListeners();
        
        // Load saved theme from localStorage
        const savedTheme = localStorage.getItem('theme') || 'light';
        document.documentElement.setAttribute('data-theme', savedTheme);
        document.querySelector('meta[name="color-scheme"]').setAttribute('content', savedTheme);
        
        // Update Pico CSS theme
        const link = document.querySelector('link[href*="pico"]');
        const newHref = link.href.replace(/pico\.[^.]+\.css/, `pico.${savedTheme}.min.css`);
        link.href = newHref;
        
        this.updateThemeState();
    }

    attachListeners() {
        this.btn?.addEventListener('click', () => this.toggleTheme());
    }

    updateThemeState() {
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
        this.btn.textContent = isDark ? 'THEME:☀️' : 'THEME:🌙';
    }

    toggleTheme() {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        document.documentElement.setAttribute('data-theme', newTheme);
        document.querySelector('meta[name="color-scheme"]').setAttribute('content', newTheme);
        
        // Update Pico CSS theme
        const link = document.querySelector('link[href*="pico"]');
        const newHref = link.href.replace(/pico\.[^.]+\.css/, `pico.${newTheme}.min.css`);
        link.href = newHref;
        
        // Save preference
        localStorage.setItem('theme', newTheme);
        
        // Update button state
        this.updateThemeState();
        
        // Dispatch event for other components
        window.dispatchEvent(new CustomEvent('_theme_changed', {
            detail: { theme: newTheme },
            bubbles: true,
            composed: true
        }));
    }
}

window.customElements.define('theme-toggle', ThemeToggle); 
