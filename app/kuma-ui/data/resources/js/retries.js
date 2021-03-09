(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["retries"],{6663:function(t,e,a){"use strict";var i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"entity-url-control"},[t.shouldDisplay?a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var i=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"}},[a("KButton",{attrs:{appearance:"secondary",size:"small"},on:{click:function(){i(t.url)}}},[a("KIcon",{attrs:{slot:"icon",icon:"externalLink"},slot:"icon"}),t._v(" "+t._s(t.copyButtonText)+" ")],1),a("div",{attrs:{slot:"content"},slot:"content"},[a("p",[t._v(t._s(t.confirmationText))])])],1)]}}],null,!1,1603401634)}):t._e()],1)},n=[],s=a("a026"),o=s["default"].extend({name:"EntityURLControl",props:{url:{type:String,required:!0},copyButtonText:{type:String,default:"Copy URL"},confirmationText:{type:String,default:"URL copied to clipboard!"}},computed:{shouldDisplay:function(){var t=this.$route.params.mesh||null;return!(!t||"all"===t)}}}),r=o,l=a("2877"),u=Object(l["a"])(r,i,n,!1,null,null,null);e["a"]=u.exports},b71f:function(t,e,a){"use strict";a.r(e);var i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"retries"},[a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,"table-data-function-text":"View","table-data-row":"name"},on:{tableAction:t.tableAction,reloadData:t.loadData}},[a("template",{slot:"additionalControls"},[this.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"retries"}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()],1),a("template",{slot:"pagination"},[a("Pagination",{attrs:{"has-previous":t.previous.length>0,"has-next":t.hasNext},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1)],2),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"}},[a("template",{slot:"tabHeader"},[a("div",[a("h3",[t._v(t._s(t.tabGroupTitle))])]),a("div",[a("EntityURLControl",{attrs:{url:t.shareUrl}})],1)]),a("template",{slot:"overview"},[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,i){return a("li",{key:i},[a("h4",[t._v(t._s(i))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])],1),a("template",{slot:"yaml"},[a("YamlView",{attrs:{title:t.entityOverviewTitle,"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})],1)],2):t._e()],1)],1)},n=[],s=(a("99af"),a("b0c0"),a("d3b7"),a("bc1e")),o=a("6663"),r=a("b912"),l=a("1d10"),u=a("1799"),c=a("2778"),p=a("251b"),y=a("ff9d"),h=a("0ada"),m={name:"Retries",metaInfo:{title:"Retries"},components:{EntityURLControl:o["a"],FrameSkeleton:l["a"],Pagination:u["a"],DataOverview:c["a"],Tabs:p["a"],YamlView:y["a"],LabelList:h["a"]},mixins:[r["a"]],data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Retries present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:[],rawEntity:null,firstEntity:null,pageSize:this.$pageSize,pageOffset:null,next:null,hasNext:!1,previous:[]}},computed:{tabGroupTitle:function(){var t=this.entity;return t?"Retry: ".concat(t.name):null},entityOverviewTitle:function(){var t=this.entity;return t?"Entity Overview for ".concat(t.name):null},shareUrl:function(){var t=this,e="".concat(window.location.origin,"#"),a=this.entity,i=function(){return t.$route.query.ns?t.$route.fullPath:"".concat(e).concat(t.$route.fullPath,"?ns=").concat(a.name)};return i()}},watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.$store.dispatch("updateSelectedTab",this.tabs[0].hash),this.$store.dispatch("updateSelectedTableRow",t.name),this.getEntity(e)},loadData:function(){var t=this;this.isLoading=!0;var e=this.$route.params.mesh||null,a=this.$route.query.ns||null,i={size:this.pageSize,offset:this.pageOffset},n=function(){return"all"===e?t.$api.getAllRetries(i):a&&a.length&&"all"!==e?t.$api.getRetry(e,a,i):t.$api.getAllRetriesFromMesh(e)},o=function(){return n().then((function(e){var i=function(){var a=e;return"total"in a?0!==a.total&&a.items&&a.items.length>0?t.sortEntities(a.items):null:a},n=i();if(i()){var o=a?n:n[0];t.firstEntity=o.name,t.getEntity(Object(s["k"])(o)),t.$store.dispatch("updateSelectedTableRow",o.name),e.next?(t.next=Object(s["e"])(e.next),t.hasNext=!0):t.hasNext=!1,t.tableData.data=a?[n]:n,t.tableDataIsEmpty=!1,t.isEmpty=!1}else t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.getEntity(null)})).catch((function(e){t.hasError=!0,t.isEmpty=!0,console.error(e)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))};o()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1;var a=this.$route.params.mesh;if(t){var i="all"===a?t.mesh:a;return this.$api.getRetry(i,t.name).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(s["f"])(t,a),e.rawEntity=Object(s["k"])(t)}else e.entity=null,e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},d=m,f=a("2877"),v=Object(f["a"])(d,i,n,!1,null,null,null);e["default"]=v.exports}}]);