import{_ as A}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-e54e3c5b.js";import{_ as Z}from"./ZoneIngressDetails.vue_vue_type_script_setup_true_lang-8bbefd03.js";import{j as V,k as $,f as B,E as L,g as O,_ as P}from"./RouteView.vue_vue_type_script_setup_true_lang-4999f19d.js";import{_ as q}from"./RouteTitle.vue_vue_type_script_setup_true_lang-65d5caa7.js";import{D as C}from"./DataOverview-92c4fef1.js";import{d as M,q as n,s as F,o as i,a as k,w as I,h as l,b as u,g as z,k as x,e as E,P as N,f as Q}from"./index-0b8ed13f.js";import{Q as p}from"./QueryParameter-70743f73.js";import"./kongponents.es-a99534bb.js";import"./DefinitionListItem-a18dd4eb.js";import"./EnvoyData-e6f53ad1.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-e187911d.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-eb706f57.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-0272b2d5.js";import"./ErrorBlock-ddeaa4b4.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-bcba5a7a.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-80496b3d.js";import"./TabsWidget-8fc36bbe.js";import"./TextWithCopyButton-45a80278.js";import"./TagList-1ea0a113.js";import"./StatusBadge-bc14a264.js";const U={class:"zoneingresses"},j={key:1,class:"kcard-stack"},G={class:"kcard-border"},K={key:0,class:"kcard-border","data-testid":"list-view-summary"},pe=M({__name:"ZoneIngressListView",props:{selectedZoneIngressName:{type:[String,null],required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(T){const c=T,g=V(),{t:v}=$(),D={title:"No Data",message:"There are no Zone Ingresses present."},_=B(),m=n(!0),d=n(null),o=n({headers:[{label:"Status",key:"status"},{label:"Name",key:"entity"}],data:[]}),r=n(null),y=n(null),b=n(c.offset);F(()=>_.getters["config/getMulticlusterStatus"],function(e){e&&w(c.offset)},{immediate:!0});async function w(e){var a;b.value=e,p.set("offset",e>0?e:null),m.value=!0,d.value=null;const t=N;try{const{items:s,next:f}=await g.getAllZoneIngressOverviews({size:t,offset:e});y.value=f,o.value.data=S(s??[]),await h({name:c.selectedZoneIngressName??((a=o.value.data[0])==null?void 0:a.entity.name)})}catch(s){o.value.data=[],r.value=null,s instanceof Error?d.value=s:console.error(s)}finally{m.value=!1}}function S(e){return e.map(t=>{const{name:a}=t,s={name:"zone-ingress-detail-view",params:{zoneIngress:a}},f=L(t.zoneIngressInsight??{});return{entity:t,detailViewRoute:s,status:f}})}async function h({name:e}){if(e===void 0){r.value=null,p.set("zoneIngress",null);return}try{r.value=await g.getZoneIngressOverview({name:e}),p.set("zoneIngress",e)}catch(t){console.error(t)}}return(e,t)=>(i(),k(P,null,{default:I(()=>[l(q,{title:u(v)("zone-ingresses.routes.items.title")},null,8,["title"]),z(),l(O,{breadcrumbs:[{to:{name:"zone-ingress-list-view"},text:u(v)("zone-ingresses.routes.items.breadcrumbs")}]},{default:I(()=>{var a;return[x("div",U,[u(_).getters["config/getMulticlusterStatus"]===!1?(i(),k(A,{key:0})):(i(),E("div",j,[x("div",G,[l(C,{"selected-entity-name":(a=r.value)==null?void 0:a.name,"page-size":u(N),"is-loading":m.value,error:d.value,"empty-state":D,"table-data":o.value,"table-data-is-empty":o.value.data.length===0,next:y.value,"page-offset":b.value,onTableAction:h,onLoadData:w},null,8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])]),z(),r.value!==null?(i(),E("div",K,[l(Z,{"zone-ingress-overview":r.value},null,8,["zone-ingress-overview"])])):Q("",!0)]))])]}),_:1},8,["breadcrumbs"])]),_:1}))}});export{pe as default};
