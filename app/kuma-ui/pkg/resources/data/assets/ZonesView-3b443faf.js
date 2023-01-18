import{k as x,c1 as K,c2 as F,c3 as H,P,bU as M,c4 as q,c5 as G,c6 as R,c7 as z,c8 as g,c9 as J,ca as j,cb as Q,cc as r,o as l,c as m,i as d,w as o,a as u,b,j as _,e as I,bV as E,F as L,cd as S}from"./index-08ba2993.js";import{Q as C}from"./QueryParameter-70743f73.js";import{_ as U}from"./CodeBlock.vue_vue_type_style_index_0_lang-e26b650c.js";import{D as Y}from"./DataOverview-1eb5b106.js";import{F as X}from"./FrameSkeleton-fa914657.js";import{_ as $}from"./LabelList.vue_vue_type_style_index_0_lang-0cdd88fc.js";import{M as ee}from"./MultizoneInfo-a0f62bfb.js";import{S as te,a as se}from"./SubscriptionHeader-4b351f66.js";import{T as ne}from"./TabsWidget-7c52524a.js";import{W as ie}from"./WarningsWidget-2a6832a7.js";import"./_commonjsHelpers-87174ba5.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-cf69250c.js";import"./ErrorBlock-21576094.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-778739a1.js";import"./StatusBadge-c118c8ba.js";import"./TagList-e8e9bfa1.js";const oe={name:"ZonesView",components:{AccordionItem:K,AccordionList:F,CodeBlock:U,DataOverview:Y,FrameSkeleton:X,LabelList:$,MultizoneInfo:ee,SubscriptionDetails:te,SubscriptionHeader:se,TabsWidget:ne,WarningsWidget:ie,KBadge:H,KButton:P,KCard:M},props:{selectedZoneName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},data(){return{isLoading:!0,isEmpty:!1,error:null,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Zones present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{key:"warnings",hideLabel:!0}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],entity:{},pageSize:q,next:null,warnings:[],subscriptionsReversed:[],codeOutput:null,zonesWithIngress:new Set,pageOffset:this.offset}},computed:{...G({multicluster:"config/getMulticlusterStatus",globalCpVersion:"config/getVersion"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1,this.tableDataIsEmpty=!1,this.init(0)}},beforeMount(){this.init(this.offset)},methods:{init(t){this.multicluster&&this.loadData(t)},filterTabs(){return this.warnings.length?this.tabs:this.tabs.filter(t=>t.hash!=="#warnings")},tableAction(t){const s=t;this.getEntity(s)},parseData(t){const{zoneInsight:s={},name:a}=t;let n="-",e="",i=!0;s.subscriptions&&s.subscriptions.length&&s.subscriptions.forEach(c=>{if(c.version&&c.version.kumaCp){n=c.version.kumaCp.version;const{kumaCpGlobalCompatible:y=!0}=c.version.kumaCp;i=y,c.config&&(e=JSON.parse(c.config).store.type)}});const h=R(s);return{...t,status:h,zoneCpVersion:n,storeType:e,hasIngress:this.zonesWithIngress.has(a)?"Yes":"No",hasEgress:this.zonesWithEgress.has(a)?"Yes":"No",withWarnings:!i}},calculateZonesWithIngress(t){const s=new Set;t.forEach(({zoneIngress:{zone:a}})=>{s.add(a)}),this.zonesWithIngress=s},calculateZonesWithEgress(t){const s=new Set;t.forEach(({zoneEgress:{zone:a}})=>{s.add(a)}),this.zonesWithEgress=s},async loadData(t){this.pageOffset=t,C.set("offset",t>0?t:null),this.isLoading=!0,this.isEmpty=!1;const s=this.$route.query.ns||null,a=this.pageSize;try{const[{data:n,next:e},{items:i},{items:h}]=await Promise.all([this.getZoneOverviews(s,a,t),z(g.getAllZoneIngressOverviews.bind(g)),z(g.getAllZoneEgressOverviews.bind(g))]);this.next=e,n.length?(this.calculateZonesWithIngress(i),this.calculateZonesWithEgress(h),this.tableData.data=n.map(this.parseData),this.tableDataIsEmpty=!1,this.isEmpty=!1,this.getEntity({name:this.selectedZoneName??n[0].name})):(this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0,this.entityIsEmpty=!0)}catch(n){n instanceof Error?this.error=n:console.error(n),this.isEmpty=!0}finally{this.isLoading=!1}},async getEntity(t){var n;this.entityIsLoading=!0,this.entityIsEmpty=!0;const s=["type","name"],a=setTimeout(()=>{this.entityIsEmpty=!0,this.entityIsLoading=!1},"500");if(t){this.entityIsEmpty=!1,this.warnings=[];try{const e=await g.getZoneOverview({name:t.name}),i=((n=e.zoneInsight)==null?void 0:n.subscriptions)??[];if(this.entity={...J(e,s),"Authentication Type":j(e)},C.set("zone",this.entity.name),this.subscriptionsReversed=Array.from(i).reverse(),i.length){const{version:h={}}=i[i.length-1],{kumaCp:c={}}=h,y=c.version||"-",{kumaCpGlobalCompatible:w=!0}=c;w||this.warnings.push({kind:Q,payload:{zoneCpVersion:y,globalCpVersion:this.globalCpVersion}}),i[i.length-1].config&&(this.codeOutput=JSON.stringify(JSON.parse(i[i.length-1].config),null,2))}}catch(e){console.error(e),this.entity={},this.entityHasError=!0,this.entityIsEmpty=!0}finally{clearTimeout(a)}}this.entityIsLoading=!1},async getZoneOverviews(t,s,a){if(t)return{data:[await g.getZoneOverview({name:t},{size:s,offset:a})],next:null};{const{items:n,next:e}=await g.getAllZoneOverviews({size:s,offset:a});return{data:n??[],next:e}}}}},ae={class:"zones"},re={class:"entity-heading"},le={key:0},ce={key:1},pe={key:2};function ue(t,s,a,n,e,i){const h=r("MultizoneInfo"),c=r("KButton"),y=r("DataOverview"),w=r("KBadge"),D=r("LabelList"),O=r("SubscriptionHeader"),A=r("SubscriptionDetails"),W=r("AccordionItem"),Z=r("AccordionList"),v=r("KCard"),N=r("CodeBlock"),T=r("WarningsWidget"),V=r("TabsWidget"),B=r("FrameSkeleton");return l(),m("div",ae,[t.multicluster===!1?(l(),d(h,{key:0})):(l(),d(B,{key:1},{default:o(()=>{var k;return[u(y,{"selected-entity-name":(k=e.entity)==null?void 0:k.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.tableDataIsEmpty,"show-warnings":e.tableData.data.some(p=>p.withWarnings),next:e.next,"page-offset":e.pageOffset,onTableAction:i.tableAction,onLoadData:i.loadData},{additionalControls:o(()=>[t.$route.query.ns?(l(),d(c,{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zones"}},{default:o(()=>[b(`
            View all
          `)]),_:1})):_("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","show-warnings","next","page-offset","onTableAction","onLoadData"]),b(),e.isEmpty===!1?(l(),d(V,{key:0,"has-error":e.error,"is-loading":e.isLoading,tabs:i.filterTabs()},{tabHeader:o(()=>[I("h1",re,`
            Zone: `+E(e.entity.name),1)]),overview:o(()=>[u(D,{"has-error":e.entityHasError,"is-loading":e.entityIsLoading,"is-empty":e.entityIsEmpty},{default:o(()=>[I("div",null,[I("ul",null,[(l(!0),m(L,null,S(e.entity,(p,f)=>(l(),m("li",{key:f},[p?(l(),m("h4",le,E(f),1)):_("",!0),b(),f==="status"?(l(),m("p",ce,[u(w,{appearance:p==="Offline"?"danger":"success"},{default:o(()=>[b(E(p),1)]),_:2},1032,["appearance"])])):(l(),m("p",pe,E(p),1))]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])]),insights:o(()=>[u(v,{"border-variant":"noBorder"},{body:o(()=>[u(Z,{"initially-open":0},{default:o(()=>[(l(!0),m(L,null,S(e.subscriptionsReversed,(p,f)=>(l(),d(W,{key:f},{"accordion-header":o(()=>[u(O,{details:p},null,8,["details"])]),"accordion-content":o(()=>[u(A,{details:p},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),config:o(()=>[e.codeOutput?(l(),d(v,{key:0,"border-variant":"noBorder"},{body:o(()=>[u(N,{id:"code-block-zone-config",language:"json",code:e.codeOutput,"is-searchable":"","query-key":"zone-config"},null,8,["code"])]),_:1})):_("",!0)]),warnings:o(()=>[u(T,{warnings:e.warnings},null,8,["warnings"])]),_:1},8,["has-error","is-loading","tabs"])):_("",!0)]}),_:1}))])}const Ce=x(oe,[["render",ue]]);export{Ce as default};
