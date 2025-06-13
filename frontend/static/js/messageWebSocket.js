class MessageWebSocket {
    constructor(userId) {
        this.userId = userId;
        this.ws = null;
        this.connect();
    }

    connect() {
        const token = authService.getToken();
        if (!token) return;
        this.ws = new WebSocket(`ws://${window.location.host}/ws/messages?token=${token}`);

        this.ws.onopen = () => {};
        this.ws.onmessage = (event) => {
            try {
                this.handleMessage(JSON.parse(event.data));
            } catch {}
        };
        this.ws.onclose = () => setTimeout(() => this.connect(), 5000);
        this.ws.onerror = () => {};
    }

    handleMessage(message) {
        switch (message.type) {
            case 'post_created':
                formHandlers.initFeed();
                break;
            case 'comment_added': {
                const postElement = document.querySelector(`[data-post-id="${message.post_id}"]`);
                if (postElement) {
                    const commentsContainer = postElement.querySelector('.comments-container');
                    commentsContainer.insertAdjacentHTML('afterbegin', uiComponents.renderComment(message.data, this.userId));
                    const span = postElement.querySelector('.comment-btn span');
                    span.textContent = parseInt(span.textContent) + 1;
                }
                break;
            }
            case 'comment_updated': {
                const updated = message.data;
                const el = document.querySelector(`[data-comment-id="${updated.id}"]`);
                if (el) {
                    el.querySelector('.comment-content p').textContent = updated.content;
                    el.querySelector('.comment-content').classList.remove('hidden');
                    el.querySelector('.edit-comment-form').classList.add('hidden');
                }
                break;
            }
            case 'comment_deleted': {
                const id = message.data.id;
                const el = document.querySelector(`[data-comment-id="${id}"]`);
                if (el) {
                    const btn = el.closest('[data-post-id]').querySelector('.comment-btn span');
                    btn.textContent = parseInt(btn.textContent) - 1;
                    el.remove();
                }
                break;
            }
            case 'post_reactions_updated': {
                const data = message.data;
                const postEl = document.querySelector(`[data-post-id="${data.id}"]`);
                if (postEl) {
                    postEl.querySelector('.like-count').textContent = data.like_count;
                    postEl.querySelector('.dislike-count').textContent = data.dislike_count;
                    const likeBtn = postEl.querySelector('.like-btn');
                    const dislikeBtn = postEl.querySelector('.dislike-btn');
                    likeBtn.classList.toggle('text-blue-600', data.user_reaction === 'like');
                    likeBtn.classList.toggle('bg-blue-50', data.user_reaction === 'like');
                    dislikeBtn.classList.toggle('text-red-600', data.user_reaction === 'dislike');
                    dislikeBtn.classList.toggle('bg-red-50', data.user_reaction === 'dislike');
                }
                break;
            }
            case 'new_message':
                // Handle new private message UI here if needed
                break;
        }
        window.dispatchEvent(new CustomEvent('wsMessage', { detail: message }));
    }

    sendMessage(receiverId, content) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
        this.ws.send(JSON.stringify({
            type: 'send_message',
            data: content,
            sender_id: this.userId,
            receiver_id: receiverId
        }));
    }

    close() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}