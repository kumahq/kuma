import{l as Z,S as F,U as W,Z as j,h as G}from"./kongponents.es-8f2ab58c.js";import{d as I,o as _,a as k,w as o,h as n,b as a,g as u,u as Y,v as H,q as p,c as g,e as J,k as m,f as N,P as L,y as U,t as X,p as ee,m as ae}from"./index-1147aef1.js";import{_ as te}from"./PolicyDetails.vue_vue_type_script_setup_true_lang-303d57fd.js";import{l as se,j as le,f as oe,k as ne,g as ie,_ as re,h as ce}from"./RouteView.vue_vue_type_script_setup_true_lang-ca499fa5.js";import{_ as pe}from"./RouteTitle.vue_vue_type_script_setup_true_lang-f1356936.js";import{D as ue}from"./DataOverview-a61bf90e.js";import{Q as $}from"./QueryParameter-70743f73.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-19dcf8fc.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-495914c1.js";import"./ErrorBlock-22abd2ad.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-c5cb903e.js";import"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-125286c8.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-99eae46c.js";import"./TextWithCopyButton-84dfce1a.js";import"./toYaml-4e00099e.js";import"./TabsWidget-37581fa3.js";import"./TagList-2dbe2543.js";import"./StatusBadge-4afb516e.js";const me=I({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(l){const t=l;return(P,T)=>(_(),k(a(F),{class:"docs-link",appearance:"outline",target:"_blank",to:t.href},{default:o(()=>[n(a(Z),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),u(`

    Documentation
  `)]),_:1},8,["to"]))}}),de=l=>(ee("data-v-d561a0c9"),l=l(),ae(),l),ye={class:"kcard-stack"},he={class:"kcard-border"},_e=de(()=>m("p",null,[m("strong",null,"Warning"),u(` This policy is experimental. If you encountered any problem please open an
                      `),m("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),fe=I({__name:"PolicyListView",props:{selectedPolicyName:{type:[String,null],required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(l){const t=l,P=se(),T=le(),d=Y(),V=H(),f=oe(),{t:C}=ne(),S=p(!0),x=p(null),w=p(null),D=p(t.offset),v=p(t.selectedPolicyName),y=p({headers:[{label:"Name",key:"entity"},{label:"Type",key:"type"}],data:[]}),M=g(()=>d.params.mesh),s=g(()=>f.state.policyTypesByPath[t.policyPath]),q=g(()=>f.state.policyTypes.map(e=>({label:e.name,value:e.path,selected:e.path===t.policyPath}))),B=g(()=>f.state.policyTypes.filter(e=>(f.state.sidebar.insights.mesh.policies[e.name]??0)===0).map(e=>e.name));R();async function R(){A(t.offset)}async function A(e){var b;D.value=e,$.set("offset",e>0?e:null),S.value=!0,x.value=null;const i=d.params.mesh,r=d.params.policyPath,c=L;try{const{items:h,next:Q}=await T.getAllPolicyEntitiesFromMesh({mesh:i,path:r},{size:c,offset:e});w.value=Q,y.value.data=z(h??[]),E({name:t.selectedPolicyName??((b=y.value.data[0])==null?void 0:b.entity.name)})}catch(h){y.value.data=[],h instanceof Error?x.value=h:console.error(h)}finally{S.value=!1}}function z(e){return e.map(i=>{const{type:r,name:c}=i,b={name:"policy-detail-view",params:{mesh:i.mesh,policyPath:d.params.policyPath,policy:c}};return{entity:i,detailViewRoute:b,type:r}})}function K(e){E({name:e.name})}function E({name:e}){v.value=e??null,$.set("policy",e??null)}function O(e){V.push({name:"policies-list-view",params:{...d.params,policyPath:e.value}})}return(e,i)=>(_(),k(re,{module:"policies"},{default:o(()=>{var r;return[n(pe,{title:a(C)("policies.routes.items.title",{name:(r=s.value)==null?void 0:r.name})},null,8,["title"]),u(),n(ie,null,{default:o(()=>[s.value?(_(),J("div",{key:0,class:U(["relative",s.value.path])},[m("div",ye,[m("div",he,[s.value.isExperimental?(_(),k(a(W),{key:0,"border-variant":"noBorder",class:"mb-4"},{body:o(()=>[n(a(j),{appearance:"warning"},{alertMessage:o(()=>[_e]),_:1})]),_:1})):N("",!0),u(),n(ue,{"selected-entity-name":v.value??void 0,"page-size":a(L),error:x.value,"is-loading":S.value,"empty-state":{title:"No Data",message:`There are no ${s.value.name} policies present.`},"table-data":y.value,"table-data-is-empty":y.value.data.length===0,next:w.value,"page-offset":D.value,onTableAction:K,onLoadData:A},{additionalControls:o(()=>[n(a(G),{label:"Policies",items:q.value,"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:O},{"item-template":o(({item:c})=>[m("span",{class:U({"policy-type-empty":B.value.includes(c.label)})},X(c.label),3)]),_:1},8,["items"]),u(),n(me,{href:`${a(P)("KUMA_DOCS_URL")}/policies/${s.value.path}/?${a(P)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"])]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"])]),u(),v.value!==null?(_(),k(te,{key:0,name:v.value,mesh:M.value,path:s.value.path,type:s.value.name},null,8,["name","mesh","path","type"])):N("",!0)])],2)):N("",!0)]),_:1})]}),_:1}))}});const Ce=ce(fe,[["__scopeId","data-v-d561a0c9"]]);export{Ce as default};
