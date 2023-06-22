import{Z as ee,t as te,S as ae}from"./kongponents.es-bba90403.js";import{d as K,q as i,o as m,a as h,w as u,n as B,g as d,b as t,f as A,u as se,s as F,a1 as W,O as ne,h as _,k as x,e as U,P as Y,l as oe,t as D}from"./index-9d631905.js";import{_ as re}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-a0429cb9.js";import{_ as le}from"./ZoneDetails.vue_vue_type_script_setup_true_lang-45ff8dae.js";import{l as ie,k as ce,j as ue,f as de,g as me,_ as pe}from"./RouteView.vue_vue_type_script_setup_true_lang-76145142.js";import{_ as fe}from"./RouteTitle.vue_vue_type_script_setup_true_lang-f639963c.js";import{D as ve}from"./DataOverview-993d8d3c.js";import{Q as O}from"./QueryParameter-70743f73.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-9125ad7e.js";import"./DefinitionListItem-ad3ab377.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-9b865501.js";import"./TabsWidget-0e0dd5da.js";import"./ErrorBlock-be40f398.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-7f9cc3f9.js";import"./TextWithCopyButton-6bd93ee0.js";import"./WarningsWidget.vue_vue_type_script_setup_true_lang-ffa4d4c0.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-255e2244.js";import"./TagList-65249449.js";import"./StatusBadge-e2897fec.js";const ge=K({__name:"DeleteResourceModal",props:{actionButtonText:{type:String,required:!1,default:"Yes, delete"},confirmationText:{type:String,required:!1,default:""},deleteFunction:{type:Function,required:!0},isVisible:{type:Boolean,required:!0},modalId:{type:String,required:!0},title:{type:String,required:!1,default:"Delete"}},emits:["cancel","delete"],setup(E,{emit:p}){const o=E,s=i(!1);async function r(){s.value=!1;try{await o.deleteFunction(),p("delete")}catch(f){console.error(f),s.value=!0}}return(f,y)=>(m(),h(t(te),{"action-button-text":o.actionButtonText,"confirmation-text":o.confirmationText,"is-visible":o.isVisible,"modal-id":o.modalId,title:o.title,type:"danger","data-testid":"delete-resource-modal",onCanceled:y[0]||(y[0]=S=>p("cancel")),onProceed:r},{"body-content":u(()=>[B(f.$slots,"body-content"),d(),s.value?(m(),h(t(ee),{key:0,class:"mt-4",appearance:"danger","is-dismissible":""},{alertMessage:u(()=>[B(f.$slots,"error")]),_:3})):A("",!0)]),_:3},8,["action-button-text","confirmation-text","is-visible","modal-id","title"]))}}),be={class:"zones"},ye={key:1,class:"kcard-stack"},_e={class:"kcard-border"},he={key:0,class:"kcard-border","data-testid":"list-view-summary"},Re=K({__name:"ZoneListView",props:{selectedZoneName:{type:[String,null],required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(E){const p=E,o=ie(),{t:s}=ce(),r=ue(),f={title:"No Data",message:"There are no Zones present."},y=se(),S=de(),N=i(!0),z=i(!1),k=i(""),T=i(null),v=i({headers:[{label:"Status",key:"status"},{label:"Name",key:"entity"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],data:[]}),g=i(null),M=i(null),V=i(p.offset);F(()=>y.params.mesh,function(){y.name==="zone-cp-list-view"&&w(0)}),F(()=>S.getters["config/getMulticlusterStatus"],function(e){e&&w(p.offset)},{immediate:!0});async function w(e){var n;V.value=e,O.set("offset",e>0?e:null),N.value=!0,T.value=null;const c=Y;try{const[{items:l,next:Z},{items:a},{items:I}]=await Promise.all([r.getAllZoneOverviews({size:c,offset:e}),W(r.getAllZoneIngressOverviews.bind(r)),W(r.getAllZoneEgressOverviews.bind(r))]);M.value=Z,v.value.data=G(l??[],a??[],I??[]),await q({name:p.selectedZoneName??((n=v.value.data[0])==null?void 0:n.entity.name)})}catch(l){g.value=null,v.value.data=[],l instanceof Error?T.value=l:console.error(l)}finally{N.value=!1}}function G(e,c,n){const l=new Set(c.map(a=>a.zoneIngress.zone)),Z=new Set(n.map(a=>a.zoneEgress.zone));return e.map(a=>{var R;const{name:I}=a,J={name:"zone-cp-detail-view",params:{zone:I}};let $="-",L="",P=!0;(((R=a.zoneInsight)==null?void 0:R.subscriptions)??[]).forEach(b=>{if(b.version&&b.version.kumaCp){$=b.version.kumaCp.version;const{kumaCpGlobalCompatible:X=!0}=b.version.kumaCp;P=X,b.config&&(L=JSON.parse(b.config).store.type)}});const H=ne(a.zoneInsight);return{entity:a,detailViewRoute:J,status:H,zoneCpVersion:$,storeType:L,hasIngress:l.has(a.name)?"Yes":"No",hasEgress:Z.has(a.name)?"Yes":"No",withWarnings:!P}})}async function q({name:e}){if(e===void 0){g.value=null,O.set("zone",null);return}try{g.value=await r.getZoneOverview({name:e}),O.set("zone",e)}catch(c){console.error(c)}}async function Q(){await r.deleteZone({name:k.value})}function C(e){var n;const c=((n=e==null?void 0:e.entity)==null?void 0:n.name)??(e==null?void 0:e.name)??"";z.value=!z.value,k.value=c}function j(){C(),w(0)}return(e,c)=>(m(),h(pe,null,{default:u(()=>[_(fe,{title:t(s)("zone-cps.routes.items.title")},null,8,["title"]),d(),_(me,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:t(s)("zone-cps.routes.items.breadcrumbs")}]},{default:u(()=>{var n;return[x("div",be,[t(S).getters["config/getMulticlusterStatus"]===!1?(m(),h(re,{key:0})):(m(),U("div",ye,[x("div",_e,[_(ve,{"selected-entity-name":(n=g.value)==null?void 0:n.name,"page-size":t(Y),"is-loading":N.value,error:T.value,"empty-state":f,"table-data":v.value,"table-data-is-empty":v.value.data.length===0,"show-warnings":v.value.data.some(l=>l.withWarnings),next:M.value,"page-offset":V.value,"show-delete-action":t(o)("KUMA_ZONE_CREATION_FLOW")==="enabled",onDeleteResource:C,onTableAction:q,onLoadData:w},oe({_:2},[t(o)("KUMA_ZONE_CREATION_FLOW")==="enabled"?{name:"additionalControls",fn:u(()=>[_(t(ae),{appearance:"creation",icon:"plus",to:{name:"zone-create-view"}},{default:u(()=>[d(`
                  Create Zone
                `)]),_:1})]),key:"0"}:void 0]),1032,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","show-warnings","next","page-offset","show-delete-action"])]),d(),g.value!==null?(m(),U("div",he,[_(le,{"zone-overview":g.value},null,8,["zone-overview"])])):A("",!0)])),d(),z.value?(m(),h(ge,{key:2,"confirmation-text":k.value,"delete-function":Q,"is-visible":z.value,"modal-id":"delete-zone-modal","action-button-text":t(s)("zones.delete.confirmModal.proceedText"),title:t(s)("zones.delete.confirmModal.title"),onCancel:C,onDelete:j},{"body-content":u(()=>[x("p",null,D(t(s)("zones.delete.confirmModal.text1",{zoneName:k.value})),1),d(),x("p",null,D(t(s)("zones.delete.confirmModal.text2")),1)]),error:u(()=>[d(D(t(s)("zones.delete.confirmModal.errorText")),1)]),_:1},8,["confirmation-text","is-visible","action-button-text","title"])):A("",!0)])]}),_:1},8,["breadcrumbs"])]),_:1}))}});export{Re as default};
