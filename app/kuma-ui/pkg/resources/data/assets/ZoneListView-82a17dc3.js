import{j as H,m as X,T as ee}from"./kongponents.es-21ce59a5.js";import{_ as te}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-dfecd6e2.js";import{_ as ae}from"./ZoneDetails.vue_vue_type_script_setup_true_lang-8a175ce5.js";import{w as se,i as ne,g as oe,e as le,x as B,v as re,A as ie,_ as ce}from"./RouteView.vue_vue_type_script_setup_true_lang-9e62d24f.js";import{_ as ue}from"./RouteTitle.vue_vue_type_script_setup_true_lang-dcec85af.js";import{D as de}from"./DataOverview-43dcbd85.js";import{d as U,q as i,o as m,a as h,w as u,n as F,g as d,b as t,f as Z,s as me,h as _,k as x,e as R,P as W,l as pe,t as I}from"./index-bdbf5b57.js";import{Q as D}from"./QueryParameter-70743f73.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-32c93dbe.js";import"./DefinitionListItem-310ce025.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-5891fe23.js";import"./TabsWidget-b5ed50b6.js";import"./ErrorBlock-b1fa7c54.js";import"./TextWithCopyButton-8088f9cb.js";import"./WarningsWidget.vue_vue_type_script_setup_true_lang-df625e99.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-ee434af6.js";import"./TagList-af54ef62.js";import"./StatusBadge-c156d8c6.js";const fe=U({__name:"DeleteResourceModal",props:{actionButtonText:{type:String,required:!1,default:"Yes, delete"},confirmationText:{type:String,required:!1,default:""},deleteFunction:{type:Function,required:!0},isVisible:{type:Boolean,required:!0},modalId:{type:String,required:!0},title:{type:String,required:!1,default:"Delete"}},emits:["cancel","delete"],setup(E,{emit:p}){const o=E,s=i(!1);async function l(){s.value=!1;try{await o.deleteFunction(),p("delete")}catch(f){console.error(f),s.value=!0}}return(f,y)=>(m(),h(t(X),{"action-button-text":o.actionButtonText,"confirmation-text":o.confirmationText,"is-visible":o.isVisible,"modal-id":o.modalId,title:o.title,type:"danger","data-testid":"delete-resource-modal",onCanceled:y[0]||(y[0]=z=>p("cancel")),onProceed:l},{"body-content":u(()=>[F(f.$slots,"body-content"),d(),s.value?(m(),h(t(H),{key:0,class:"mt-4",appearance:"danger","is-dismissible":""},{alertMessage:u(()=>[F(f.$slots,"error")]),_:3})):Z("",!0)]),_:3},8,["action-button-text","confirmation-text","is-visible","modal-id","title"]))}}),ve={class:"zones"},ge={key:1,class:"stack"},be={class:"kcard-border"},ye={key:0,class:"kcard-border","data-testid":"list-view-summary"},Le=U({__name:"ZoneListView",props:{selectedZoneName:{type:[String,null],required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(E){const p=E,o=se(),{t:s}=ne(),l=oe(),f={title:"No Data",message:"There are no Zones present."},y=le(),z=i(!0),w=i(!1),k=i(""),T=i(null),v=i({headers:[{label:"Status",key:"status"},{label:"Name",key:"entity"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],data:[]}),g=i(null),O=i(null),M=i(p.offset);me(()=>y.getters["config/getMulticlusterStatus"],function(e){e&&N(p.offset)},{immediate:!0});async function N(e){var n;M.value=e,D.set("offset",e>0?e:null),z.value=!0,T.value=null;const c=W;try{const[{items:r,next:C},{items:a},{items:A}]=await Promise.all([l.getAllZoneOverviews({size:c,offset:e}),B(l.getAllZoneIngressOverviews.bind(l)),B(l.getAllZoneEgressOverviews.bind(l))]);O.value=C,v.value.data=Y(r??[],a??[],A??[]),await V({name:p.selectedZoneName??((n=v.value.data[0])==null?void 0:n.entity.name)})}catch(r){g.value=null,v.value.data=[],r instanceof Error?T.value=r:console.error(r)}finally{z.value=!1}}function Y(e,c,n){const r=new Set(c.map(a=>a.zoneIngress.zone)),C=new Set(n.map(a=>a.zoneEgress.zone));return e.map(a=>{var $;const{name:A}=a,G={name:"zone-cp-detail-view",params:{zone:A}};let q="-",L="",P=!0;((($=a.zoneInsight)==null?void 0:$.subscriptions)??[]).forEach(b=>{if(b.version&&b.version.kumaCp){q=b.version.kumaCp.version;const{kumaCpGlobalCompatible:J=!0}=b.version.kumaCp;P=J,b.config&&(L=JSON.parse(b.config).store.type)}});const Q=re(a.zoneInsight);return{entity:a,detailViewRoute:G,status:Q,zoneCpVersion:q,storeType:L,hasIngress:r.has(a.name)?"Yes":"No",hasEgress:C.has(a.name)?"Yes":"No",withWarnings:!P}})}async function V({name:e}){if(e===void 0){g.value=null,D.set("zone",null);return}try{g.value=await l.getZoneOverview({name:e}),D.set("zone",e)}catch(c){console.error(c)}}async function K(){await l.deleteZone({name:k.value})}function S(e){var n;const c=((n=e==null?void 0:e.entity)==null?void 0:n.name)??(e==null?void 0:e.name)??"";w.value=!w.value,k.value=c}function j(){S(),N(0)}return(e,c)=>(m(),h(ce,null,{default:u(()=>[_(ue,{title:t(s)("zone-cps.routes.items.title")},null,8,["title"]),d(),_(ie,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:t(s)("zone-cps.routes.items.breadcrumbs")}]},{default:u(()=>{var n;return[x("div",ve,[t(y).getters["config/getMulticlusterStatus"]===!1?(m(),h(te,{key:0})):(m(),R("div",ge,[x("div",be,[_(de,{"selected-entity-name":(n=g.value)==null?void 0:n.name,"page-size":t(W),"is-loading":z.value,error:T.value,"empty-state":f,"table-data":v.value,"table-data-is-empty":v.value.data.length===0,"show-warnings":v.value.data.some(r=>r.withWarnings),next:O.value,"page-offset":M.value,"show-delete-action":t(o)("KUMA_ZONE_CREATION_FLOW")==="enabled",onDeleteResource:S,onTableAction:V,onLoadData:N},pe({_:2},[t(o)("KUMA_ZONE_CREATION_FLOW")==="enabled"?{name:"additionalControls",fn:u(()=>[_(t(ee),{appearance:"creation",icon:"plus",to:{name:"zone-create-view"}},{default:u(()=>[d(`
                  Create Zone
                `)]),_:1})]),key:"0"}:void 0]),1032,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","show-warnings","next","page-offset","show-delete-action"])]),d(),g.value!==null?(m(),R("div",ye,[_(ae,{"zone-overview":g.value},null,8,["zone-overview"])])):Z("",!0)])),d(),w.value?(m(),h(fe,{key:2,"confirmation-text":k.value,"delete-function":K,"is-visible":w.value,"modal-id":"delete-zone-modal","action-button-text":t(s)("zones.delete.confirmModal.proceedText"),title:t(s)("zones.delete.confirmModal.title"),onCancel:S,onDelete:j},{"body-content":u(()=>[x("p",null,I(t(s)("zones.delete.confirmModal.text1",{zoneName:k.value})),1),d(),x("p",null,I(t(s)("zones.delete.confirmModal.text2")),1)]),error:u(()=>[d(I(t(s)("zones.delete.confirmModal.errorText")),1)]),_:1},8,["confirmation-text","is-visible","action-button-text","title"])):Z("",!0)])]}),_:1},8,["breadcrumbs"])]),_:1}))}});export{Le as default};
