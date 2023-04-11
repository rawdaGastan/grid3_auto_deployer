<template>
  <v-footer class="bg-grey-darken-4 pa-0" absolute app height="75">
    <v-row class="text-center">
      <v-col cols="12">
        All rights reserved &copy; {{ new Date().getFullYear() }} —
        <a
          href="https://www.codescalers-egypt.com"
          target="_blank"
          class="font-weight-bold text-white text-decoration-none"
          >CodeScalers Egypt</a
        >
        <div class="d-inline-flex ml-5 social">
          <a href="mailto:info@codescalers.com" class="mr-5 rounded-circle"
            ><font-awesome-icon :icon="['fas', 'envelope']" color="white"
          /></a>
          <a
            href="https://www.linkedin.com/company/codescalers-egypt/mycompany/"
            target="_blank"
            class="rounded-circle"
            ><font-awesome-icon :icon="['fab', 'linkedin-in']" color="white"
          /></a>
        </div>
      </v-col>
    </v-row>
  </v-footer>
</template>

<style>
.social a {
  padding: 10px;
  border: 1px solid #fff;
  height: 40px;
  width: 40px;
  line-height: 20px;
}
</style>
