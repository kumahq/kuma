import{D as x,cn as K,co as F,U as H,e as R,Q as M,P,cp as q,cq as G,k as m,cr as C,s as U,cs as J,ct as j,o as l,j as g,c as h,w as o,a as p,b as L,z as _,l as d,t as E,F as S,n as z,i as a}from"./index.3a3d021f.js";import{g as Y}from"./tableDataUtils.db5b56fc.js";import{_ as Q}from"./CodeBlock.vue_vue_type_style_index_0_lang.1ff82bc9.js";import{D as X}from"./DataOverview.529ae0fd.js";import{_ as $}from"./EntityURLControl.vue_vue_type_script_setup_true_lang.c0a3cdda.js";import{F as ee}from"./FrameSkeleton.0eaca3db.js";import{_ as te}from"./LabelList.vue_vue_type_style_index_0_lang.7ae57614.js";import{M as se}from"./MultizoneInfo.85e59ea2.js";import{S as ne,a as ie}from"./SubscriptionHeader.10e71be7.js";import{T as oe}from"./TabsWidget.8d50c7e4.js";import{W as ae}from"./WarningsWidget.9029457d.js";import"./_commonjsHelpers.f037b798.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.ed08f409.js";import"./ErrorBlock.e950a812.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.fa633f42.js";import"./TagList.4e52fad1.js";const re={name:"ZonesView",components:{AccordionItem:K,AccordionList:F,CodeBlock:Q,DataOverview:X,EntityURLControl:$,FrameSkeleton:ee,LabelList:te,MultizoneInfo:se,SubscriptionDetails:ne,SubscriptionHeader:ie,TabsWidget:oe,WarningsWidget:ae,KBadge:H,KButton:R,KCard:M},data(){return{isLoading:!0,isEmpty:!1,error:null,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Zones present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{key:"warnings",hideLabel:!0}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],entity:{},pageSize:P,next:null,warnings:[],subscriptionsReversed:[],codeOutput:null,zonesWithIngress:new Set}},computed:{...q({multicluster:"config/getMulticlusterStatus",globalCpVersion:"config/getVersion"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1,this.tableDataIsEmpty=!1,this.init()}},beforeMount(){this.init()},methods:{init(){this.multicluster&&this.loadData()},filterTabs(){return this.warnings.length?this.tabs:this.tabs.filter(s=>s.hash!=="#warnings")},tableAction(s){const t=s;this.getEntity(t)},parseData(s){const{zoneInsight:t={},name:n}=s;let u="-",e="",r=!0;return t.subscriptions&&t.subscriptions.length&&t.subscriptions.forEach(i=>{if(i.version&&i.version.kumaCp){u=i.version.kumaCp.version;const{kumaCpGlobalCompatible:y=!0}=i.version.kumaCp;r=y,i.config&&(e=JSON.parse(i.config).store.type)}}),{...s,status:G(t).status,zoneCpVersion:u,storeType:e,hasIngress:this.zonesWithIngress.has(n)?"Yes":"No",hasEgress:this.zonesWithEgress.has(n)?"Yes":"No",withWarnings:!r}},calculateZonesWithIngress(s){const t=new Set;s.forEach(({zoneIngress:{zone:n}})=>{t.add(n)}),this.zonesWithIngress=t},calculateZonesWithEgress(s){const t=new Set;s.forEach(({zoneEgress:{zone:n}})=>{t.add(n)}),this.zonesWithEgress=t},async loadData(s="0"){this.isLoading=!0,this.isEmpty=!1;const t=this.$route.query.ns||null;try{const[{data:n,next:u},{items:e},{items:r}]=await Promise.all([Y({getSingleEntity:m.getZoneOverview.bind(m),getAllEntities:m.getAllZoneOverviews.bind(m),size:this.pageSize,offset:s,query:t}),C({callEndpoint:m.getAllZoneIngressOverviews.bind(m)}),C({callEndpoint:m.getAllZoneEgressOverviews.bind(m)})]);this.next=u,n.length?(this.calculateZonesWithIngress(e),this.calculateZonesWithEgress(r),this.tableData.data=n.map(this.parseData),this.tableDataIsEmpty=!1,this.isEmpty=!1,this.getEntity({name:n[0].name})):(this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0,this.entityIsEmpty=!0)}catch(n){n instanceof Error?error.value=n:console.error(n),this.isEmpty=!0}finally{this.isLoading=!1}},async getEntity(s){var u,e;this.entityIsLoading=!0,this.entityIsEmpty=!0;const t=["type","name"],n=setTimeout(()=>{this.entityIsEmpty=!0,this.entityIsLoading=!1},"500");if(s){this.entityIsEmpty=!1,this.warnings=[];try{const r=await m.getZoneOverview({name:s.name}),i=(e=(u=r.zoneInsight)==null?void 0:u.subscriptions)!=null?e:[];if(this.entity={...U(r,t),"Authentication Type":J(r)},this.subscriptionsReversed=Array.from(i).reverse(),i.length){const{version:y={}}=i[i.length-1],{kumaCp:b={}}=y,I=b.version||"-",{kumaCpGlobalCompatible:v=!0}=b;v||this.warnings.push({kind:j,payload:{zoneCpVersion:I,globalCpVersion:this.globalCpVersion}}),i[i.length-1].config&&(this.codeOutput=JSON.stringify(JSON.parse(i[i.length-1].config),null,2))}}catch(r){console.error(r),this.entity={},this.entityHasError=!0,this.entityIsEmpty=!0}finally{clearTimeout(n)}}this.entityIsLoading=!1}}},le={class:"zones"},ce=d("span",{class:"custom-control-icon"}," \u2190 ",-1),pe={key:0},me={key:1},ue={key:2};function ge(s,t,n,u,e,r){const i=a("MultizoneInfo"),y=a("KButton"),b=a("DataOverview"),I=a("EntityURLControl"),v=a("KBadge"),D=a("LabelList"),A=a("SubscriptionHeader"),W=a("SubscriptionDetails"),O=a("AccordionItem"),T=a("AccordionList"),w=a("KCard"),Z=a("CodeBlock"),B=a("WarningsWidget"),V=a("TabsWidget"),N=a("FrameSkeleton");return l(),g("div",le,[s.multicluster===!1?(l(),h(i,{key:0})):(l(),h(N,{key:1},{default:o(()=>{var k;return[p(b,{"selected-entity-name":(k=e.entity)==null?void 0:k.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.tableDataIsEmpty,"show-warnings":e.tableData.data.some(c=>c.withWarnings),next:e.next,onTableAction:r.tableAction,onLoadData:t[0]||(t[0]=c=>r.loadData(c))},{additionalControls:o(()=>[s.$route.query.ns?(l(),h(y,{key:0,class:"back-button",appearance:"primary",to:{name:"zones"}},{default:o(()=>[ce,L(" View All ")]),_:1})):_("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","show-warnings","next","onTableAction"]),e.isEmpty===!1?(l(),h(V,{key:0,"has-error":e.error,"is-loading":e.isLoading,tabs:r.filterTabs(),"initial-tab-override":"overview"},{tabHeader:o(()=>[d("div",null,[d("h3",null," Zone: "+E(e.entity.name),1)]),d("div",null,[p(I,{name:e.entity.name},null,8,["name"])])]),overview:o(()=>[p(D,{"has-error":e.entityHasError,"is-loading":e.entityIsLoading,"is-empty":e.entityIsEmpty},{default:o(()=>[d("div",null,[d("ul",null,[(l(!0),g(S,null,z(e.entity,(c,f)=>(l(),g("li",{key:f},[c?(l(),g("h4",pe,E(f),1)):_("",!0),f==="status"?(l(),g("p",me,[p(v,{appearance:c==="Offline"?"danger":"success"},{default:o(()=>[L(E(c),1)]),_:2},1032,["appearance"])])):(l(),g("p",ue,E(c),1))]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])]),insights:o(()=>[p(w,{"border-variant":"noBorder"},{body:o(()=>[p(T,{"initially-open":0},{default:o(()=>[(l(!0),g(S,null,z(e.subscriptionsReversed,(c,f)=>(l(),h(O,{key:f},{"accordion-header":o(()=>[p(A,{details:c},null,8,["details"])]),"accordion-content":o(()=>[p(W,{details:c},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),config:o(()=>[e.codeOutput?(l(),h(w,{key:0,"border-variant":"noBorder"},{body:o(()=>[p(Z,{id:"code-block-zone-config",language:"json",code:e.codeOutput,"is-searchable":"","query-key":"zone-config"},null,8,["code"])]),_:1})):_("",!0)]),warnings:o(()=>[p(B,{warnings:e.warnings},null,8,["warnings"])]),_:1},8,["has-error","is-loading","tabs"])):_("",!0)]}),_:1}))])}const Ae=x(re,[["render",ge]]);export{Ae as default};
