package stats

// NOTE: The structs in this package were generated by getting responses from ES, using
// this handy tool at https://mholt.github.io/json-to-go/ and making adjustments

type Cluster struct {
	Response struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_nodes"`
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Timestamp   int64  `json:"timestamp"`
	Status      string `json:"status"`
	Indices     struct {
		Count  int `json:"count"`
		Shards struct {
			Total       int     `json:"total"`
			Primaries   int     `json:"primaries"`
			Replication float64 `json:"replication"`
			Index       struct {
				Shards struct {
					Min int     `json:"min"`
					Max int     `json:"max"`
					Avg float64 `json:"avg"`
				} `json:"shards"`
				Primaries struct {
					Min int     `json:"min"`
					Max int     `json:"max"`
					Avg float64 `json:"avg"`
				} `json:"primaries"`
				Replication struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
					Avg float64 `json:"avg"`
				} `json:"replication"`
			} `json:"index"`
		} `json:"shards"`
		Docs struct {
			Count   int `json:"count"`
			Deleted int `json:"deleted"`
		} `json:"docs"`
		Store struct {
			SizeInBytes     int64 `json:"size_in_bytes"`
			ReservedInBytes int   `json:"reserved_in_bytes"`
		} `json:"store"`
		Fielddata struct {
			MemorySizeInBytes int `json:"memory_size_in_bytes"`
			Evictions         int `json:"evictions"`
		} `json:"fielddata"`
		QueryCache struct {
			MemorySizeInBytes int `json:"memory_size_in_bytes"`
			TotalCount        int `json:"total_count"`
			HitCount          int `json:"hit_count"`
			MissCount         int `json:"miss_count"`
			CacheSize         int `json:"cache_size"`
			CacheCount        int `json:"cache_count"`
			Evictions         int `json:"evictions"`
		} `json:"query_cache"`
		Completion struct {
			SizeInBytes int `json:"size_in_bytes"`
		} `json:"completion"`
		Segments struct {
			Count                     int `json:"count"`
			MemoryInBytes             int `json:"memory_in_bytes"`
			TermsMemoryInBytes        int `json:"terms_memory_in_bytes"`
			StoredFieldsMemoryInBytes int `json:"stored_fields_memory_in_bytes"`
			TermVectorsMemoryInBytes  int `json:"term_vectors_memory_in_bytes"`
			NormsMemoryInBytes        int `json:"norms_memory_in_bytes"`
			PointsMemoryInBytes       int `json:"points_memory_in_bytes"`
			DocValuesMemoryInBytes    int `json:"doc_values_memory_in_bytes"`
			IndexWriterMemoryInBytes  int `json:"index_writer_memory_in_bytes"`
			VersionMapMemoryInBytes   int `json:"version_map_memory_in_bytes"`
			FixedBitSetMemoryInBytes  int `json:"fixed_bit_set_memory_in_bytes"`
			MaxUnsafeAutoIDTimestamp  int `json:"max_unsafe_auto_id_timestamp"`
			FileSizes                 struct {
			} `json:"file_sizes"`
		} `json:"segments"`
		Mappings struct {
			FieldTypes []struct {
				Name       string `json:"name"`
				Count      int    `json:"count"`
				IndexCount int    `json:"index_count"`
			} `json:"field_types"`
		} `json:"mappings"`
		Analysis struct {
			CharFilterTypes    []interface{} `json:"char_filter_types"`
			TokenizerTypes     []interface{} `json:"tokenizer_types"`
			FilterTypes        []interface{} `json:"filter_types"`
			AnalyzerTypes      []interface{} `json:"analyzer_types"`
			BuiltInCharFilters []interface{} `json:"built_in_char_filters"`
			BuiltInTokenizers  []interface{} `json:"built_in_tokenizers"`
			BuiltInFilters     []interface{} `json:"built_in_filters"`
			BuiltInAnalyzers   []interface{} `json:"built_in_analyzers"`
		} `json:"analysis"`
	} `json:"indices"`
	Nodes struct {
		Count struct {
			Total               int `json:"total"`
			CoordinatingOnly    int `json:"coordinating_only"`
			Data                int `json:"data"`
			Ingest              int `json:"ingest"`
			Master              int `json:"master"`
			RemoteClusterClient int `json:"remote_cluster_client"`
		} `json:"count"`
		Versions []string `json:"versions"`
		Os       struct {
			AvailableProcessors int `json:"available_processors"`
			AllocatedProcessors int `json:"allocated_processors"`
			Names               []struct {
				Count int `json:"count"`
			} `json:"names"`
			PrettyNames []struct {
				Count int `json:"count"`
			} `json:"pretty_names"`
			Mem struct {
				TotalInBytes int64 `json:"total_in_bytes"`
				FreeInBytes  int64 `json:"free_in_bytes"`
				UsedInBytes  int64 `json:"used_in_bytes"`
				FreePercent  int   `json:"free_percent"`
				UsedPercent  int   `json:"used_percent"`
			} `json:"mem"`
		} `json:"os"`
		Process struct {
			CPU struct {
				Percent int `json:"percent"`
			} `json:"cpu"`
			OpenFileDescriptors struct {
				Min int `json:"min"`
				Max int `json:"max"`
				Avg int `json:"avg"`
			} `json:"open_file_descriptors"`
		} `json:"process"`
		Jvm struct {
			MaxUptimeInMillis int64 `json:"max_uptime_in_millis"`
			Mem               struct {
				HeapUsedInBytes int64 `json:"heap_used_in_bytes"`
				HeapMaxInBytes  int64 `json:"heap_max_in_bytes"`
			} `json:"mem"`
			Threads int `json:"threads"`
		} `json:"jvm"`
		Fs struct {
			TotalInBytes     int64 `json:"total_in_bytes"`
			FreeInBytes      int64 `json:"free_in_bytes"`
			AvailableInBytes int64 `json:"available_in_bytes"`
		} `json:"fs"`
		NetworkTypes struct {
			TransportTypes struct {
				Netty4 int `json:"netty4"`
			} `json:"transport_types"`
			HTTPTypes struct {
				FilterJetty int `json:"filter-jetty"`
			} `json:"http_types"`
		} `json:"network_types"`
		DiscoveryTypes struct {
			Zen int `json:"zen"`
		} `json:"discovery_types"`
		PackagingTypes []struct {
			Flavor string `json:"flavor"`
			Type   string `json:"type"`
			Count  int    `json:"count"`
		} `json:"packaging_types"`
		Ingest struct {
			NumberOfPipelines int `json:"number_of_pipelines"`
			ProcessorStats    struct {
			} `json:"processor_stats"`
		} `json:"ingest"`
	} `json:"nodes"`
}
