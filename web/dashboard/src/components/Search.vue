<template>
  <div class="hello">
    <h2>Search package behaviour</h2>
    <b-form-group label="Type">
      <b-form-radio-group v-model="type">
        <b-form-radio value="file">File</b-form-radio>
        <b-form-radio value="command">Command</b-form-radio>
        <b-form-radio value="ip">IP address</b-form-radio>
      </b-form-radio-group>
    </b-form-group>
    <b-form-input
      v-model="search"
      placeholder="Path component (e.g. '.ssh')"
    ></b-form-input>
    <b-table :fields="fields" :items="results">
      <template #cell(ecosystemPkg)="data">
        {{ data.item.Package.Ecosystem }}/{{ data.item.Package.Name }}
      </template>
      <template #cell(version)="data">
        {{ data.item.Package.Version }}
        <router-link :to="getPackageLink(data.item.Package)">
          <b-icon icon="link"></b-icon>
        </router-link>
      </template>
    </b-table>
  </div>
</template>

<script>
export default {
  name: 'Search',
  props: {},
  methods: {
    async makeRequest() {
      const curId = this.requestId;
      const backend = process.env.VUE_APP_BACKEND;
      const response = await fetch(
          `${backend}/query`, {
            method: 'POST',
            body: JSON.stringify({
              'type': this.type,
              'search': this.search,
            }),
          });

      const results = await response.json();
      if (curId != this.requestId) {
        // A newer request superseded this one.
        return;
      }

      this.results = results.packages;
    },

    getPackageLink(pkg) {
      return {
        name: 'package',
        query: {
          ecosystem: pkg.Ecosystem,
          name: pkg.Name,
          version: pkg.Version,
        }
      }
    }
  },
  data() {
    return {
      requestId: 0,
      changeId: 0,
      type: 'file',
      search: '',
      results: [],
      fields: [
        { key: 'ecosystemPkg', label: 'Package' },
        'version',
      ],
    };
  },
  watch: {
    search() {
      const curId = ++this.requestId;
      setTimeout(
        () => {
          if (curId == this.requestId) {
            this.makeRequest();
          }
        },
        600
      );
    },
  },
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
