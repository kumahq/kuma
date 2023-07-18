import{_ as S}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-3a7fd74c.js";import{_ as V}from"./ZoneIngressDetails.vue_vue_type_script_setup_true_lang-88a342d8.js";import{g as Z,i as B,e as L,v as O,A as P,_ as $}from"./RouteView.vue_vue_type_script_setup_true_lang-0ac8938c.js";import{_ as q}from"./RouteTitle.vue_vue_type_script_setup_true_lang-cccbfca9.js";import{D as C}from"./DataOverview-30ce4833.js";import{d as M,q as n,s as F,o as i,a as I,w as k,h as l,b as u,g as z,k as x,e as A,P as N,f as Q}from"./index-f0e2f93b.js";import{Q as p}from"./QueryParameter-70743f73.js";import"./kongponents.es-d49ba82d.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-b0190284.js";import"./DefinitionListItem-8aa6d45d.js";import"./EnvoyData-aca02a9a.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-58df6732.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-8302aaa3.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-fd5f25bc.js";import"./ErrorBlock-3bc373a3.js";import"./TabsWidget-444de6c7.js";import"./TextWithCopyButton-c830f326.js";import"./TagList-7e09ae10.js";import"./StatusBadge-9ddf65b2.js";const U={class:"zoneingresses"},G={key:1,class:"kcard-stack"},K={class:"kcard-border"},R={key:0,class:"kcard-border","data-testid":"list-view-summary"},fe=M({__name:"ZoneIngressListView",props:{selectedZoneIngressName:{type:[String,null],required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(T){const c=T,g=Z(),{t:v}=B(),D={title:"No Data",message:"There are no Zone Ingresses present."},_=L(),m=n(!0),d=n(null),o=n({headers:[{label:"Status",key:"status"},{label:"Name",key:"entity"}],data:[]}),r=n(null),y=n(null),b=n(c.offset);F(()=>_.getters["config/getMulticlusterStatus"],function(e){e&&w(c.offset)},{immediate:!0});async function w(e){var a;b.value=e,p.set("offset",e>0?e:null),m.value=!0,d.value=null;const t=N;try{const{items:s,next:f}=await g.getAllZoneIngressOverviews({size:t,offset:e});y.value=f,o.value.data=E(s??[]),await h({name:c.selectedZoneIngressName??((a=o.value.data[0])==null?void 0:a.entity.name)})}catch(s){o.value.data=[],r.value=null,s instanceof Error?d.value=s:console.error(s)}finally{m.value=!1}}function E(e){return e.map(t=>{const{name:a}=t,s={name:"zone-ingress-detail-view",params:{zoneIngress:a}},f=O(t.zoneIngressInsight??{});return{entity:t,detailViewRoute:s,status:f}})}async function h({name:e}){if(e===void 0){r.value=null,p.set("zoneIngress",null);return}try{r.value=await g.getZoneIngressOverview({name:e}),p.set("zoneIngress",e)}catch(t){console.error(t)}}return(e,t)=>(i(),I($,null,{default:k(()=>[l(q,{title:u(v)("zone-ingresses.routes.items.title")},null,8,["title"]),z(),l(P,{breadcrumbs:[{to:{name:"zone-ingress-list-view"},text:u(v)("zone-ingresses.routes.items.breadcrumbs")}]},{default:k(()=>{var a;return[x("div",U,[u(_).getters["config/getMulticlusterStatus"]===!1?(i(),I(S,{key:0})):(i(),A("div",G,[x("div",K,[l(C,{"selected-entity-name":(a=r.value)==null?void 0:a.name,"page-size":u(N),"is-loading":m.value,error:d.value,"empty-state":D,"table-data":o.value,"table-data-is-empty":o.value.data.length===0,next:y.value,"page-offset":b.value,onTableAction:h,onLoadData:w},null,8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])]),z(),r.value!==null?(i(),A("div",R,[l(V,{"zone-ingress-overview":r.value},null,8,["zone-ingress-overview"])])):Q("",!0)]))])]}),_:1},8,["breadcrumbs"])]),_:1}))}});export{fe as default};
