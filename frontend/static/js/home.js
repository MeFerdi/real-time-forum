
function renderHome() {
    document.getElementById('app').innerHTML = `
        <div class="flex items-center justify-center min-h-screen">
            <div class="bg-white p-8 rounded shadow-md w-full max-w-md space-y-4 text-center">
                <h2 class="text-2xl font-bold mb-4">Welcome Home!</h2>
                <p class="mb-4">You are logged in.</p>
                <button id="logout-btn" class="bg-red-500 text-white p-2 rounded">Logout</button>
            </div>
        </div>
    `;
    document.getElementById('logout-btn').onclick = function() {
        if (window.authService) {
            window.authService.logout();
        }
        window.location.hash = '#login';
        if (typeof renderLogin === 'function') {
            renderLogin();
        }
    };
}

// Make renderHome globally accessible
window.renderHome = renderHome;
