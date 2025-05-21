// Clase EditableList
class EditableList extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.items = [];
        this._renderInitial();
    }

    _renderInitial() {
        this.shadowRoot.innerHTML = `
            <style>
                .container {
                    padding: 20px;
                    font-family: Arial, sans-serif;
                }
                
                table {
                    width: 100%;
                    border-collapse: collapse;
                    margin-bottom: 20px;
                }
                
                th, td {
                    padding: 12px;
                    text-align: left;
                    border-bottom: 1px solid #ddd;
                }

                .edit-form {
                    display: none;
                    margin-top: 20px;
                    padding: 20px;
                    background-color: #f9f9f9;
                    border-radius: 4px;
                }

                .edit-form.active {
                    display: block;
                }

                .form-group {
                    margin-bottom: 15px;
                }

                label {
                    display: block;
                    margin-bottom: 5px;
                }

                input {
                    width: 100%;
                    padding: 8px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }

                button {
                    padding: 8px 16px;
                    margin: 0 4px;
                    border: none;
                    border-radius: 4px;
                    cursor: pointer;
                }

                .edit-btn {
                    background-color: #4CAF50;
                    color: white;
                }

                .delete-btn {
                    background-color: #f44336;
                    color: white;
                }

                .submit-btn {
                    background-color: #2196F3;
                    color: white;
                }
            </style>
            
            <div class="container">
                <table>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Nombre</th>
                            <th>Descripción</th>
                            <th>Precio</th>
                            <th>Acciones</th>
                        </tr>
                    </thead>
                    <tbody id="items-list"></tbody>
                </table>

                <div class="edit-form" id="editForm">
                    <h3>Editar Producto</h3>
                    <form id="edit-product-form">
                        <input type="hidden" id="edit-id">
                        <div class="form-group">
                            <label for="edit-nombre">Nombre:</label>
                            <input type="text" id="edit-nombre" required>
                        </div>
                        <div class="form-group">
                            <label for="edit-descripcion">Descripción:</label>
                            <input type="text" id="edit-descripcion" required>
                        </div>
                        <div class="form-group">
                            <label for="edit-precio">Precio:</label>
                            <input type="number" id="edit-precio" step="0.01" required>
                        </div>
                        <button type="submit" class="submit-btn">Actualizar</button>
                        <button type="button" class="cancel-btn" onclick="this._hideEditForm()">Cancelar</button>
                    </form>
                </div>

                <div class="create-form">
                    <h3>Crear Nuevo Producto</h3>
                    <form id="create-form">
                        <div class="form-group">
                            <label for="nombre">Nombre:</label>
                            <input type="text" id="nombre" required>
                        </div>
                        <div class="form-group">
                            <label for="descripcion">Descripción:</label>
                            <input type="text" id="descripcion" required>
                        </div>
                        <div class="form-group">
                            <label for="precio">Precio:</label>
                            <input type="number" id="precio" step="0.01" required>
                        </div>
                        <button type="submit" class="submit-btn">Crear Producto</button>
                    </form>
                </div>
            </div>
        `;

        this._bindEvents();
    }

    _bindEvents() {
        const createForm = this.shadowRoot.getElementById('create-form');
        const editForm = this.shadowRoot.getElementById('edit-product-form');

        createForm.addEventListener('submit', (e) => this._handleCreate(e));
        editForm.addEventListener('submit', (e) => this._handleEditSubmit(e));
    }

    _renderItems() {
        const tbody = this.shadowRoot.getElementById('items-list');
        tbody.innerHTML = '';
        
        this.items.forEach(item => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${item.id}</td>
                <td>${item.nombre}</td>
                <td>${item.descripcion}</td>
                <td>$${item.precio}</td>
                <td>
                    <button class="edit-btn" data-id="${item.id}">✏️ Editar</button>
                    <button class="delete-btn" data-id="${item.id}">🗑️ Eliminar</button>
                </td>
            `;
            
            const editBtn = tr.querySelector('.edit-btn');
            const deleteBtn = tr.querySelector('.delete-btn');
            
            editBtn.addEventListener('click', () => this._showEditForm(item));
            deleteBtn.addEventListener('click', () => this._handleDelete(item.id));
            
            tbody.appendChild(tr);
        });
    }

    _showEditForm(item) {
        const editForm = this.shadowRoot.getElementById('editForm');
        const idInput = this.shadowRoot.getElementById('edit-id');
        const nombreInput = this.shadowRoot.getElementById('edit-nombre');
        const descripcionInput = this.shadowRoot.getElementById('edit-descripcion');
        const precioInput = this.shadowRoot.getElementById('edit-precio');

        idInput.value = item.id;
        nombreInput.value = item.nombre;
        descripcionInput.value = item.descripcion;
        precioInput.value = item.precio;

        editForm.classList.add('active');
    }

    _hideEditForm() {
        const editForm = this.shadowRoot.getElementById('editForm');
        editForm.classList.remove('active');
    }

    _handleEditSubmit(e) {
        e.preventDefault();
        const id = this.shadowRoot.getElementById('edit-id').value;
        const editedItem = {
            id: id,
            nombre: this.shadowRoot.getElementById('edit-nombre').value,
            descripcion: this.shadowRoot.getElementById('edit-descripcion').value,
            precio: parseFloat(this.shadowRoot.getElementById('edit-precio').value)
        };

        this.dispatchEvent(new CustomEvent('item-edit', {
            bubbles: true,
            composed: true,
            detail: { item: editedItem }
        }));

        this._hideEditForm();
    }

    setData(items) {
        this.items = items;
        this._renderItems();
    }
}

// Registrar el componente
customElements.define('editable-list', EditableList);