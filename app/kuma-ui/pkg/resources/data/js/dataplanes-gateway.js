"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[897],{41930:function(t,e,a){a.r(e),a.d(e,{default:function(){return u}});var s=function(){var t=this,e=t._self._c;return e("ServicesView",{attrs:{"route-name":"external-services",name:"External Services","tab-headers":t.tabHeaders}})},n=[],i=a(28744),r={name:"ExternalServices",metaInfo:{title:"External Services"},components:{ServicesView:i.Z},data(){return{tabHeaders:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Address",key:"address"},{label:"TLS",key:"tlsEnabled"}]}}},l=r,o=a(1001),h=(0,o.Z)(l,s,n,!1,null,null,null),u=h.exports},96486:function(t,e,a){a.r(e),a.d(e,{default:function(){return u}});var s=function(){var t=this,e=t._self._c;return e("div",{staticClass:"gateway-dataplanes"},[e("DataplanesView",t._b({},"DataplanesView",t.$data,!1))],1)},n=[],i=a(5274),r={name:"GatewayDataplanes",metaInfo:{title:"Gateway Data plane proxies"},components:{DataplanesView:i.Z},data(){return{dataplaneApiParams:{gateway:!0},emptyStateMsg:"There are no Gateway data plane proxies present."}}},l=r,o=a(1001),h=(0,o.Z)(l,s,n,!1,null,null,null),u=h.exports},4933:function(t,e,a){a.r(e),a.d(e,{default:function(){return u}});var s=function(){var t=this,e=t._self._c;return e("ServicesView",{attrs:{"route-name":"internal-services",name:"Internal Services","tab-headers":t.tabHeaders}})},n=[],i=a(28744),r={name:"InternalServices",metaInfo:{title:"Internal Services"},components:{ServicesView:i.Z},data(){return{tabHeaders:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Data plane proxies: Online / Total",key:"totalOnline"}]}}},l=r,o=a(1001),h=(0,o.Z)(l,s,n,!1,null,null,null),u=h.exports},28744:function(t,e,a){a.d(e,{Z:function(){return E}});var s=function(){var t=this,e=t._self._c;return e("div",[e("FrameSkeleton",[e("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[t.$route.query.ns?e("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:t.routeName}}},[e("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])},[t._v(" > ")]),!1===t.isEmpty?e("TabsWidget",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[e("div",[e("h3",[t._v(t._s(t.name)+": "+t._s(t.entity.name))])]),e("div",[e("EntityURLControl",{attrs:{name:t.entity.name,mesh:t.entity.mesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[e("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[e("div",[e("ul",t._l(t.entity,(function(a,s){return e("li",{key:s},[e("h4",[t._v(t._s(s))]),e("p",[t._v(" "+t._s(a)+" ")])])})),0)])])]},proxy:!0},{key:"yaml",fn:function(){return[e("YamlView",{attrs:{lang:"yaml","has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0}],null,!1,1914485537)}):t._e()],1)],1)},n=[],i=a(53419),r=a(17463),l=a(87673),o={methods:{sortEntities(t){const e=t.sort(((t,e)=>t.name>e.name||t.name===e.name&&t.mesh>e.mesh?1:-1));return e}}},h=a(88523),u=a(84855),m=a(56882),c=a(7001),y=a(59316),d=a(33561),p=a(45689),v={name:"ServicesView",components:{EntityURLControl:l.Z,FrameSkeleton:u.Z,DataOverview:m.Z,TabsWidget:c.Z,YamlView:y.Z,LabelList:d.Z},mixins:[h.Z,o],props:{routeName:{type:String,required:!0},name:{type:String,default:""},tabHeaders:{type:Array,required:!0}},data(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:`There are not ${this.name} present.`},tableData:{headers:this.tabHeaders,data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:{},rawEntity:{},pageSize:p.NR,next:null}},computed:{formattedRawEntity(){const t=this.formatForCLI(this.rawEntity);return t}},watch:{$route(t,e){this.init()}},beforeMount(){this.init()},methods:{getAllServices(t){return"Internal Services"===this.name?r.Z.getAllServiceInsights(t):r.Z.getAllExternalServices(t)},getService(t,e,a){return"Internal Services"===this.name?r.Z.getServiceInsight({mesh:t,name:e},a):r.Z.getExternalService({mesh:t,name:e},a)},getServiceFromMesh(t){return"Internal Services"===this.name?r.Z.getAllServiceInsightsFromMesh({mesh:t}):r.Z.getAllExternalServicesFromMesh({mesh:t})},parseData(t){if("Internal Services"===this.name){const{dataplanes:e={}}=t,{online:a=0,total:s=0}=e;switch(t.totalOnline=`${a} / ${s}`,t.status){case"online":t.status=p.ku;break;case"partially_degraded":t.status=p.hP;break;case"offline":default:t.status=p.cz}return t}const{networking:e={}}=t,{tls:a={}}=e;return t.address=e.address,t.tlsEnabled=a.enabled?"Enabled":"Disabled",t},init(){this.loadData()},goToPreviousPage(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction(t){const e=t;this.getEntity(e)},loadData(t="0"){this.isLoading=!0;const e=this.$route.params.mesh||null,a=this.$route.query.ns||null,s={size:this.pageSize,offset:t},n=()=>"all"===e?this.getAllServices(s):a&&a.length&&"all"!==e?this.getService(e,a,s):this.getServiceFromMesh(e),r=()=>n().then((t=>{const e=()=>{const e=t;return"total"in e?0!==e.total&&e.items&&e.items.length>0?this.sortEntities(e.items):null:e},s=e();if(e()){const e=a?s:s[0];this.getEntity((0,i.RV)(e)),this.tableData.data=a?[s]:s,this.next=Boolean(t.next),this.tableData.data=this.tableData.data.map(this.parseData),this.tableDataIsEmpty=!1,this.isEmpty=!1}else this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0,this.getEntity(null)})).catch((t=>{this.hasError=!0,this.isEmpty=!0,console.error(t)})).finally((()=>{setTimeout((()=>{this.isLoading=!1}),"500")}));r()},getEntity(t){this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1;const e=this.$route.params.mesh;if(t&&null!==t){const a="all"===e?t.mesh:e;return this.getService(a,t.name).then((t=>{if(t){const e=["type","name","mesh"];this.entity=(0,i.wy)(t,e),this.rawEntity=(0,i.RV)(t)}else this.entity={},this.entityIsEmpty=!0})).catch((t=>{this.entityHasError=!0,console.error(t)})).finally((()=>{setTimeout((()=>{this.entityIsLoading=!1}),"500")}))}setTimeout((()=>{this.entityIsEmpty=!0,this.entityIsLoading=!1}),"500")}}},g=v,f=a(1001),b=(0,f.Z)(g,s,n,!1,null,null,null),E=b.exports},88523:function(t,e,a){var s=a(73570),n=a.n(s);e["Z"]={methods:{formatForCLI(t,e='" | kumactl apply -f -'){const a='echo "',s=n()(t);return`${a}${s}${e}`}}}}}]);