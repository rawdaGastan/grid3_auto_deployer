<template>
  <v-footer class="d-flex bg-grey-darken-4 pa-0" absolute app height="75">
    <v-row class="text-center">
      <v-divider />
      <v-col cols="12">
        All rights reserved &copy; {{ new Date().getFullYear() }} —
        <a
          href="https://www.codescalers-egypt.com"
          target="_blank"
          class="font-weight-bold text-white text-decoration-none"
          >CodeScalers Egypt</a
        >
        <div class="d-inline-flex ml-5 social">
          <a href="mailto:info@codescalers.com" class="mr-md-5 rounded-circle"
            ><v-icon>mdi-email</v-icon></a
          >
          <a
            href="https://www.linkedin.com/company/codescalers-egypt/"
            target="_blank"
            class="rounded-circle"
            ><v-icon>mdi-linkedin</v-icon></a
          >
        </div>
      </v-col>
    </v-row>
  </v-footer>
</template>

<style>
.social a {
  padding: 10px;
  border: 1px solid #474747;
  line-height: 0px;
  color: #fff;
}

@media only screen and (max-width: 960px) {
  .social a {
    border: none;
  }
}
</style>
