<script>
export default {
    name: 'Modal',
    methods: {
        close() {
            this.$emit('close');
        },
        accept() {
            this.$emit('accept');
        },
        decline() {
            this.$emit('decline');
        },
    },
};
</script>

<template>
    <transition name="modal-fade">
        <div class="modal-backdrop">
            <div class="modal" role="dialog" aria-labelledby="modalTitle" aria-describedby="modalDescription" >
                    <header class="modal-header" id="modalTitle" sticky>
                        <slot name="header">
                            <h4 style="color: #217DBB;">Terms and Condition</h4>
                        </slot>
                        <button type="button" class="btn-close" @click="close" aria-label="Close modal" style="color: #217DBB;">
                            x
                        </button>
                    </header>


                <section class="modal-body" id="modalDescription">
                    <slot name="body">

                        Lorem ipsum dolor sit amet consectetur adipisicing elit. Modi commodi earum tenetur. Asperiores
                        dolorem
                        placeat ab nobis iusto culpa, autem molestias molestiae quidem pariatur. Debitis beatae expedita nam
                        facere perspiciatis. Lorem ipsum dolor sit amet consectetur adipisicing elit. Repellendus ducimus
                        cupiditate rerum officiis consequuntur laborum doloremque quaerat ipsa voluptates, nobis nam quis
                        nulla
                        ullam at corporis, similique ratione quasi illo! Lorem ipsum dolor sit amet consectetur adipisicing
                        elit. Modi commodi earum tenetur. Asperiores
                        dolorem
                        placeat ab nobis iusto culpa, autem molestias molestiae quidem pariatur. Debitis beatae expedita nam
                        facere perspiciatis. Lorem ipsum dolor sit amet consectetur adipisicing elit. Repellendus ducimus
                        cupiditate rerum officiis consequuntur laborum doloremque quaerat ipsa voluptates, nobis nam quis
                        nulla
                        ullam at corporis, similique ratione quasi illo! Lorem ipsum dolor sit amet consectetur adipisicing
                        elit. Modi commodi earum tenetur. Asperiores
                        dolorem
                        placeat ab nobis iusto culpa, autem molestias molestiae quidem pariatur. Debitis beatae expedita nam
                        facere perspiciatis. Lorem ipsum dolor sit amet consectetur adipisicing elit. Repellendus ducimus
                        cupiditate rerum officiis consequuntur laborum doloremque quaerat ipsa voluptates, nobis nam quis
                        nulla
                        ullam at corporis, similique ratione quasi illo! Lorem ipsum dolor sit amet consectetur adipisicing
                        elit. Modi commodi earum tenetur. Asperiores
                        dolorem
                        placeat ab nobis iusto culpa, autem molestias molestiae quidem pariatur. Debitis beatae expedita nam
                        facere perspiciatis. Lorem ipsum dolor sit amet consectetur adipisicing elit. Repellendus ducimus
                        cupiditate rerum officiis consequuntur laborum doloremque quaerat ipsa voluptates, nobis nam quis
                        nulla
                        ullam at corporis, similique ratione quasi illo!
                    </slot>
                </section>

                <footer class="modal-footer">
                    <slot name="footer">
                        <div class="d-flex my-3">


                            <v-btn @click="decline"  class=" text-capitalize mx-auto" style="background-color: #FF5252 !important;color:white;">
                                Decline
                            </v-btn>

                            <v-btn @click="accept" variant="flat" color="primary"
                                class=" text-capitalize mx-auto bg-primary">
                                Accept
                            </v-btn>




                        </div>
                    </slot>

                </footer>
            </div>
        </div>
    </transition>
</template>

<style>
.modal-backdrop {
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    background-color: rgba(0, 0, 0, 0.3);
    display: flex;
    justify-content: center;
    align-items: center;
}

.modal {
    background: #FFFFFF;
    box-shadow: 2px 2px 20px 1px;
    overflow-x: auto;
    display: flex;
    flex-direction: column;
    max-width: 50%;
    max-height: 40%;
}

.modal-header,
.modal-footer {
    padding: 15px;
    display: flex;
}

.modal-header {
    position: relative;
    border-bottom: 1px solid #eeeeee;
    justify-content: space-between;
}

.modal-footer {
    border-top: 1px solid #eeeeee;
    flex-direction: column;
}

.modal-body {
    position: relative;
    padding: 20px 10px;
}

.btn-close {
    position: absolute;
    top: 0;
    right: 0;
    border: none;
    font-size: 20px;
    padding: 10px;
    cursor: pointer;
    font-weight: bold;
    background: transparent;
}





.modal-fade-enter,
.modal-fade-leave-to {
    opacity: 0;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
    transition: opacity .5s ease;
}
</style>
