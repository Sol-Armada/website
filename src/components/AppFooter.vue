<template>
  <!-- <v-footer height="40" app> -->
  <v-sheet>
    <v-row class="justify-center text-center">
      <v-col cols="12">
        <a v-for="item in items" :key="item.title" :href="item.href" :title="item.title"
          class="d-inline-block mx-2 social-link" rel="noopener noreferrer" target="_blank">
          <v-icon :icon="item.icon" :size="item.icon === '$vuetify' ? 24 : 16" />
        </a>
        <div class="text-caption text-disabled">
          &copy; 2016-{{ (new Date()).getFullYear() }} <span class="d-none d-sm-inline-block">Sol Armada</span>
        </div>
      </v-col>
    </v-row>
  </v-sheet>
  <!-- </v-footer> -->
</template>

<script setup>
const items = [
  {
    title: 'Sol Armada GitHub',
    icon: `mdi:mdi-github`,
    href: 'https://github.com/sol-armada',
  },
  {
    title: 'Sol Armada Discord',
    icon: `fa fa-brands fa-discord`,
    href: 'https://community.vuetifyjs.com/',
  },
  {
    title: 'Sol Armada YouTube',
    icon: `fa fa-brands fa-youtube`,
    href: 'https://www.youtube.com/@Sol-Armada',
  },
]
</script>

<style scoped lang="sass">
  .social-link :deep(.v-icon)
    color: rgba(var(--v-theme-on-background), var(--v-disabled-opacity))
    text-decoration: none
    transition: .2s ease-in-out

    &:hover
      color: rgba(25, 118, 210, 1)
</style>
