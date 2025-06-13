const utils = {
    setupImagePreview(inputId, previewContainerId, previewImgId, removeBtnId) {
        const imageInput = document.getElementById(inputId);
        const previewContainer = document.getElementById(previewContainerId);
        const previewImg = document.getElementById(previewImgId);
        const removeButton = document.getElementById(removeBtnId);

        if (!imageInput || !previewContainer || !previewImg || !removeButton) return;

        imageInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = (ev) => {
                    previewImg.src = ev.target.result;
                    previewContainer.classList.remove('hidden');
                };
                reader.readAsDataURL(file);
            }
        });

        removeButton.addEventListener('click', () => {
            imageInput.value = '';
            previewContainer.classList.add('hidden');
            previewImg.src = '';
        });
    },

    formatDate(dateStr) {
        const date = new Date(dateStr);
        if (isNaN(date)) return '';
        return date.toLocaleString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
};

export { utils };