// Initialize the application
document.addEventListener('DOMContentLoaded', async () => {
    // Initialize router
    router.init();

    // Initialize views and check authentication
    await window.views.init();
    
    // Load categories for the filter
    const categories = await API.getCategories();
    if (categories.success) {
        const categoryFilter = document.getElementById('categoryFilter');
        categories.data.forEach(category => {
            const option = document.createElement('option');
            option.value = category.id;
            option.textContent = category.name;
            categoryFilter.appendChild(option);
        });
    }
});