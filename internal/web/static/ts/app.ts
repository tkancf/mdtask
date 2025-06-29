// App TypeScript
document.addEventListener('DOMContentLoaded', (): void => {
    // Add any interactive features here
    initializeEventListeners();
});

function initializeEventListeners(): void {
    // Initialize delete confirmations
    const deleteButtons = document.querySelectorAll<HTMLButtonElement>('.delete-btn');
    deleteButtons.forEach(button => {
        button.addEventListener('click', (e: Event) => {
            if (!confirm('Are you sure you want to delete this task?')) {
                e.preventDefault();
            }
        });
    });

    // Initialize form validations
    const taskForms = document.querySelectorAll<HTMLFormElement>('.task-form');
    taskForms.forEach(form => {
        form.addEventListener('submit', (e: Event) => {
            const titleInput = form.querySelector<HTMLInputElement>('input[name="title"]');
            if (titleInput && !titleInput.value.trim()) {
                e.preventDefault();
                alert('Title is required');
            }
        });
    });

    // Auto-resize textareas
    const textareas = document.querySelectorAll<HTMLTextAreaElement>('textarea.auto-resize');
    textareas.forEach(textarea => {
        textarea.addEventListener('input', () => {
            textarea.style.height = 'auto';
            textarea.style.height = `${textarea.scrollHeight}px`;
        });
        // Trigger initial resize
        textarea.dispatchEvent(new Event('input'));
    });

    // Tag click handler for filtering
    const tagLinks = document.querySelectorAll<HTMLAnchorElement>('.tag-link');
    tagLinks.forEach(link => {
        link.addEventListener('click', (e: Event) => {
            e.preventDefault();
            const tag = link.dataset.tag;
            if (tag) {
                window.location.href = `/?tags=${encodeURIComponent(tag)}`;
            }
        });
    });
}