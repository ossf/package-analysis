<template>
  <div>
    <h3>{{$route.query.ecosystem}}/{{$route.query.name}} @ {{$route.query.version}}</h3>
    <h4>Files touched</h4>
    <b-table :items="data.Files">
    </b-table>
    <h4>Commands</h4>
    <b-table :items="data.Commands">
    </b-table>
    <h4>IPs</h4>
    <b-table :items="getIpItems(data.IPs)">
    </b-table>
  </div>
</template>

<script>
export default {
  name: 'Package',
  components: {},
  methods: {
    async makeRequest() {
      const resultsUrl = process.env.VUE_APP_ANALYSIS_RESULTS;
      const response = await fetch(
          resultsUrl + '/' +
          this.$route.query.ecosystem + '/' +
          this.$route.query.name + '/' +
          this.$route.query.version + '.json')
      this.data = await response.json();
      if (this.data.Files) {
        this.data.Files.sort((a, b) => {
          if (a.Path == b.Path) return 0;
          if (a.Path < b.Path) return -1;
          return 1;
        });
      }
    },
    getIpItems(ips) {
      const items = [];
      for (let ip of ips) {
        items.push({ip});
      }
      return items;
    }
  },
  data() {
    return {
      data: {},
    }
  },
  watch: {
    '$route'() {
      this.makeRequest();
    }
  },

  async mounted() {
    await this.makeRequest();
  }
}
</script>
