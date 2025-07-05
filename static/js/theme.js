const themeSwitcher = document.getElementById('theme-switcher');
const doc = document.documentElement;

// Function to set the theme
function setTheme(theme) {
    doc.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
}

// Event listener for the button
themeSwitcher.addEventListener('click', (e) => {
    e.preventDefault(); // Prevent the link from navigating
    const currentTheme = localStorage.getItem('theme') || 'dark';
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
});

// Set initial theme on page load
const savedTheme = localStorage.getItem('theme') || 'dark';
setTheme(savedTheme);