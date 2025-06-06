function renderHome() {

    window.logout = async function() {
        try {
            // Clear local storage and any auth tokens
            localStorage.removeItem('token');
            
            // Close WebSocket connection if it exists
            if (window.authService && window.authService.ws) {
                window.authService.ws.close();
            }

            // Call the backend logout endpoint if needed
            const response = await fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) {
                console.error('Logout failed:', response.statusText);
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            // Always redirect to login page and clean up state
            if (window.authService) {
                window.authService.token = null;
                window.authService.userId = null;
                window.authService.nickname = '';
            }
            window.location.hash = '#login';
            if (typeof renderLogin === 'function') {
                renderLogin();
            }
        }
    };

    // Initial render
    render();
}

// Make renderHome globally accessible
window.renderHome = renderHome;