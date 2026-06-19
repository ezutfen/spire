<template>
  <div>


    <table
      class="eq-table bordered mb-0"
    >
      <thead class="eq-table-floating-header">
      <tr>
        <td class="font-weight-bold p-3">Options</td>
      </tr>
      </thead>

      <tbody>
      <tr>
        <td>
          <eq-checkbox
            :fade-when-not-true="true"
            class="d-inline-block mr-3"
            :true-value="true"
            :false-value="false"
            v-model="launcher.updateOpcodesOnStart"
            @change="saveLauncherOptions()"
          />
          Update Server Patches (Opcodes) On Start)
        </td>
      </tr>
      <tr>
        <td>
          <eq-checkbox
            :fade-when-not-true="true"
            class="d-inline-block mr-3"
            :true-value="true"
            :false-value="false"
            v-model="launcher.runSharedMemory"
            @change="saveLauncherOptions()"
          />
          Run Shared Memory (Recommended)
        </td>
      </tr>
      <tr>
        <td>
          <eq-checkbox
            :fade-when-not-true="true"
            class="d-inline-block mr-3"
            :true-value="true"
            :false-value="false"
            v-model="launcher.runUcs"
            @change="saveLauncherOptions()"
          />
          Run UCS (Optional)
        </td>
      </tr>
      <tr>
        <td>
          <eq-checkbox
            :fade-when-not-true="true"
            class="d-inline-block mr-3"
            :true-value="true"
            :false-value="false"
            v-model="launcher.runLoginserver"
            @change="saveLauncherOptions()"
          />
          Run Loginserver (Optional)
        </td>
      </tr>
      <tr>
        <td>
          <eq-checkbox
            :fade-when-not-true="true"
            class="d-inline-block mr-3"
            :true-value="true"
            :false-value="false"
            v-model="launcher.runQueryServ"
            @change="saveLauncherOptions()"
          />
          Run QueryServ (Optional)
        </td>
      </tr>
      </tbody>
    </table>


    <div class="mb-3 mt-4">
      Static Zones

      <div class="mt-3">
        <div v-if="staticZones.length > 0" class="d-flex flex-wrap gap-2 mb-3">
          <span
            v-for="tag in staticZones"
            :key="tag"
            class="badge bg-success d-inline-flex align-items-center"
            :title="tag"
          >
            <span>{{ tag }}</span>
            <button
              type="button"
              class="btn btn-link text-white p-0 ml-2"
              aria-label="Remove zone"
              @click="removeStaticZone(tag)"
            >&times;</button>
          </span>
        </div>

        <select
          v-model="selectedZoneToAdd"
          class="form-select"
          :disabled="availableOptions.length === 0"
          @change="addStaticZone"
        >
          <option disabled value="">Choose a zone...</option>
          <option
            v-for="option in availableOptions"
            :key="option"
            :value="option"
          >
            {{ option }}
          </option>
        </select>
      </div>


      <div class="mt-3">
        <div>
          Min Zone Processes (Ready)
        </div>

        <div>
          <p class="text-muted">
            This is the number of zones that Spire will attempt to keep running <b>without</b> players. For example: if
            you have 10 zones with players in it and your minZoneProcesses is set to 10, you will have 20 total zones
            booted.
          </p>
        </div>
        <input
          type="number"
          class="form-control"
          v-model.number="launcher.minZoneProcesses"
          @change="saveLauncherOptions()"
        />
      </div>

      <div class="mt-3">
        <div>
          Days to keep log files (7 days default)
        </div>

        <div>
          <p class="text-muted">
            Files older than this will be deleted periodically. Set to -1 to disable.
          </p>
        </div>
        <input
          type="number"
          class="form-control"
          v-model.number="launcher.deleteLogFilesOlderThanDays"
          @change="saveLauncherOptions()"
        />
      </div>

    </div>
  </div>
</template>

<script>
import {Zones}    from "@/app/zones";
import {SpireApi} from "@/app/api/spire-api";
import EqCheckbox from "@/components/eq-ui/EQCheckbox.vue";

export default {
  name: 'LauncherOptions',
  components: { EqCheckbox },
  props: ['launcherConfig'],
  data() {
    return {
      launcher: {
        runSharedMemory: false,
        runLoginserver: false,
        runQueryServ: false,
        runUcs: true,
        updateOpcodesOnStart: true,
        staticZones: ""
      },

      staticZones: [],
      availableZoneOptions: [],
      selectedZoneToAdd: "",
    }
  },
  async created() {
    this.applyLauncherConfig(this.launcherConfig)

    // zone options
    let options = []
    const zones = await Zones.getZones()
    for (let z of zones) {
      options.push(z.short_name)
    }
    this.availableZoneOptions = options
  },
  watch: {
    launcherConfig(newValue) {
      this.applyLauncherConfig(newValue)
    }
  },
  computed: {
    availableOptions() {
      return this.availableZoneOptions.filter(opt => this.staticZones.indexOf(opt) === -1)
    }
  },
  methods: {
    applyLauncherConfig(newValue) {
      this.launcher = newValue || this.launcher

      if (this.launcher.staticZones && this.launcher.staticZones.length > 0) {
        this.staticZones = this.launcher.staticZones.split(",")
      } else {
        this.staticZones = []
      }

      if (typeof this.launcher.updateOpcodesOnStart === 'undefined') {
        this.launcher.updateOpcodesOnStart = true
      }

      if (typeof this.launcher.deleteLogFilesOlderThanDays !== 'undefined' && this.launcher.deleteLogFilesOlderThanDays === 0) {
        this.launcher.deleteLogFilesOlderThanDays = 7
      }

      if (typeof this.launcher.runUcs === 'undefined') {
        this.launcher.runUcs = true
      }
    },
    addStaticZone() {
      if (!this.selectedZoneToAdd || this.staticZones.includes(this.selectedZoneToAdd)) {
        this.selectedZoneToAdd = ""
        return
      }

      this.staticZones = [...this.staticZones, this.selectedZoneToAdd]
      this.selectedZoneToAdd = ""
      this.saveLauncherOptions()
    },
    removeStaticZone(tag) {
      this.staticZones = this.staticZones.filter((zone) => zone !== tag)
      this.saveLauncherOptions()
    },
    saveLauncherOptions() {
      setTimeout(async () => {
        this.launcher.staticZones = this.staticZones.join(",")

        try {
          await SpireApi.v1().post('admin/launcherconfig', this.launcher)
        } catch (e) {
          console.log(e)
        }

      }, 100)
    }
  }
}
</script>

<style scoped>

</style>
