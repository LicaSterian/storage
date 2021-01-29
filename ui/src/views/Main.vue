<template>
  <b-container>
    <header class="mt-3">
      <b-row>
        <b-col><h3 class="pull-left">Storage</h3></b-col>
        <b-col
          ><b-button
            class="float-right"
            variant="success"
            @click="showUploadFileModal = true"
            >+ Upload File</b-button
          ></b-col
        >
      </b-row>

      <b-modal v-model="showUploadFileModal">
        <input type="file" id="file" ref="file" @change="handleFileUpload()" />

        <template #modal-footer>
          <div class="w-100">
            <b-button
              variant="primary"
              class="float-right ml-2"
              @click="submitFile()"
            >
              Upload
            </b-button>
            <b-button
              variant="secondary"
              class="float-right"
              @click="showUploadFileModal = false"
            >
              Cancel
            </b-button>
          </div>
        </template>
      </b-modal>
    </header>
    <div class="mt-3">
      <b-row>
        <b-col sm="3">
          <b-form-input v-model="name" placeholder="Search Name"></b-form-input>
        </b-col>
        <b-col sm="1">
          <b-button variant="primary" @click="onClickSearch"
            ><b-icon icon="search"></b-icon
          ></b-button>
        </b-col>
      </b-row>
      <br />
      <b-table striped bordered hover :fields="fields" :items="files">
        <template v-slot:cell(name)="{ item }"
          ><a :href="`${apiUrl}/file/${item.id}`">{{ item.name }}</a></template
        >
        <template v-slot:cell(created_at)="{ value }">
          {{ datetime(value) }}
        </template>
        <template v-slot:cell(size)="{ value }">
          {{ filesize(value) }}
        </template>
        <template v-slot:cell(actions)="data">
          <b-icon
            icon="x-circle"
            variant="danger"
            @click="deleteFile(data.item.id)"
          ></b-icon>
        </template>
        <template #table-caption
          >Total Files: <strong>{{ totalFiles }}</strong></template
        >
      </b-table>
      <b-pagination
        v-model="page"
        :total-rows="totalFiles"
        :per-page="perPage"
        align="center"
        @change="onPaginationChange"
      ></b-pagination>
    </div>
  </b-container>
</template>

<script>
/* eslint-disable */
import axios from "axios";
import moment from "moment";
import filesize from "filesize";

export default {
  name: "Main",
  components: {},
  created() {
    this.apiUrl = process.env.VUE_APP_API_URL;
    this.getPage(1);
  },
  data() {
    return {
      apiUrl: "",

      showUploadFileModal: false,
      file: "",

      fields: [
        { key: "name", label: "Name", sortable: true },
        { key: "created_at", label: "Created At", sortable: true },
        { key: "size", label: "Size", sortable: true },
        { key: "actions", label: "Actions" },
      ],
      files: [],
      totalFiles: 0,
      addedFileId: "",
      page: 1,
      perPage: 10,

      name: "",
    };
  },
  computed: {
    filters() {
      let result = [];
      if (this.name != "") {
        result.push({
          field: "name",
          operation: "$like",
          value: this.name,
        });
      }
      return result;
    },
  },
  methods: {
    onClickSearch() {
      this.getPage(1);
    },
    getPage(page) {
      this.page = page;
      axios
        .post(`${this.apiUrl}/files`, {
          page,
          perPage: this.perPage,
          filters: this.filters,
          sortBy: "created_at",
          sortAsc: true,
        })
        .then((res) => {
          this.totalFiles = res.data.data.total;
          if (this.addedFileId === "") {
            this.files = res.data.data.rows;
          } else {
            this.files = res.data.data.rows.map((file) => {
              if (file.id === this.addedFileId) {
                return { ...file, _rowVariant: "info" };
              }
              return file;
            });
          }
        })
        .catch(() => {
          console.log("error");
        });
    },
    handleFileUpload() {
      this.file = this.$refs.file.files[0];
    },
    submitFile() {
      let formData = new FormData();
      formData.append("document", this.file);
      axios
        .post(`${this.apiUrl}/file/upload`, formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        })
        .then((res) => {
          console.log("SUCCESS!!", res.data);
          if (res.data.success) {
            let file = res.data.file;
            this.addedFileId = file.id;
            this.getPage(1);
            setTimeout(this.removeRowInfo, 3000);
          }
        })
        .catch(function() {
          console.log("FAILURE!!");
        })
        .finally(() => {
          this.showUploadFileModal = false;
        });
    },
    removeRowInfo() {
      let files = [];
      this.files.forEach((file) => {
        if (file._rowVariant !== "info") {
          files.push(file);
        } else {
          files.push({
            id: file.id,
            name: file.name,
            created_at: file.created_at,
            size: file.size,
          });
        }
      });
      this.files = files;
    },
    deleteFile(id) {
      console.log("delete file", id);
      axios
        .delete(`${this.apiUrl}/file/${id}`)
        .then((res) => {
          if (res.data.success) {
            this.getPage(1);
          }
          console.log("delete successfull", res.data.message);
        })
        .catch((err) => {
          console.log("delete error", err);
        })
        .finally(() => {});
    },
    onPaginationChange(page) {
      this.getPage(page);
    },
    datetime(timestamp) {
      return moment(timestamp, "X").format("DD/MM/YYYY LTS");
    },
    filesize(value) {
      return filesize(value);
    },
  },
};
</script>

<style></style>
