import{l as Z,S as F,U as W,Z as j,h as G}from"./kongponents.es-f7b55123.js";import{d as I,o as _,a as P,w as n,h as i,b as t,g as m,u as Y,v as H,q as u,c as g,s as J,e as X,k as d,f as N,P as L,z as U,t as ee,p as ae,m as te}from"./index-a4a530d1.js";import{_ as se}from"./PolicyDetails.vue_vue_type_script_setup_true_lang-2a0ba52f.js";import{l as le,j as oe,f as ne,k as ie,g as re,_ as ce,h as pe}from"./RouteView.vue_vue_type_script_setup_true_lang-8e6a23b5.js";import{_ as ue}from"./RouteTitle.vue_vue_type_script_setup_true_lang-f4fc2caf.js";import{D as me}from"./DataOverview-7c85e051.js";import{Q as $}from"./QueryParameter-70743f73.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-8a43cc10.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-ddebee78.js";import"./ErrorBlock-cc9ab0db.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-2c479ce1.js";import"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-6b64ed3d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-15a6fd20.js";import"./TextWithCopyButton-4f241d23.js";import"./toYaml-4e00099e.js";import"./TabsWidget-a42020cd.js";import"./TagList-76ef6044.js";import"./StatusBadge-6b87699d.js";const de=I({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(l){const a=l;return(k,T)=>(_(),P(t(F),{class:"docs-link",appearance:"outline",target:"_blank",to:a.href},{default:n(()=>[i(t(Z),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),m(`

    Documentation
  `)]),_:1},8,["to"]))}}),ye=l=>(ae("data-v-ed39ec55"),l=l(),te(),l),he={class:"kcard-stack"},_e={class:"kcard-border"},fe=ye(()=>d("p",null,[d("strong",null,"Warning"),m(` This policy is experimental. If you encountered any problem please open an
                      `),d("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),ve=I({__name:"PolicyListView",props:{selectedPolicyName:{type:[String,null],required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(l){const a=l,k=le(),T=oe(),o=Y(),V=H(),f=ne(),{t:C}=ie(),S=u(!0),w=u(null),D=u(null),A=u(a.offset),v=u(a.selectedPolicyName),y=u({headers:[{label:"Name",key:"entity"},{label:"Type",key:"type"}],data:[]}),M=g(()=>o.params.mesh),s=g(()=>f.state.policyTypesByPath[a.policyPath]),q=g(()=>f.state.policyTypes.map(e=>({label:e.name,value:e.path,selected:e.path===a.policyPath}))),z=g(()=>f.state.policyTypes.filter(e=>(f.state.sidebar.insights.mesh.policies[e.name]??0)===0).map(e=>e.name));J(()=>o.params.mesh,function(){o.name===a.policyPath&&x(0)}),B();async function B(){x(a.offset)}async function x(e){var b;A.value=e,$.set("offset",e>0?e:null),S.value=!0,w.value=null;const r=o.params.mesh,c=o.params.policyPath,p=L;try{const{items:h,next:Q}=await T.getAllPolicyEntitiesFromMesh({mesh:r,path:c},{size:p,offset:e});D.value=Q,y.value.data=R(h??[]),E({name:a.selectedPolicyName??((b=y.value.data[0])==null?void 0:b.entity.name)})}catch(h){y.value.data=[],h instanceof Error?w.value=h:console.error(h)}finally{S.value=!1}}function R(e){return e.map(r=>{const{type:c,name:p}=r,b={name:"policy-detail-view",params:{mesh:r.mesh,policyPath:o.params.policyPath,policy:p}};return{entity:r,detailViewRoute:b,type:c}})}function K(e){E({name:e.name})}function E({name:e}){v.value=e??null,$.set("policy",e??null)}function O(e){V.push({name:"policies-list-view",params:{...o.params,policyPath:e.value}})}return(e,r)=>(_(),P(ce,{module:"policies"},{default:n(()=>{var c;return[i(ue,{title:t(C)("policies.routes.items.title",{name:(c=s.value)==null?void 0:c.name})},null,8,["title"]),m(),i(re,null,{default:n(()=>[s.value?(_(),X("div",{key:0,class:U(["relative",s.value.path])},[d("div",he,[d("div",_e,[s.value.isExperimental?(_(),P(t(W),{key:0,"border-variant":"noBorder",class:"mb-4"},{body:n(()=>[i(t(j),{appearance:"warning"},{alertMessage:n(()=>[fe]),_:1})]),_:1})):N("",!0),m(),i(me,{"selected-entity-name":v.value??void 0,"page-size":t(L),error:w.value,"is-loading":S.value,"empty-state":{title:"No Data",message:`There are no ${s.value.name} policies present.`},"table-data":y.value,"table-data-is-empty":y.value.data.length===0,next:D.value,"page-offset":A.value,onTableAction:K,onLoadData:x},{additionalControls:n(()=>[i(t(G),{label:"Policies",items:q.value,"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:O},{"item-template":n(({item:p})=>[d("span",{class:U({"policy-type-empty":z.value.includes(p.label)})},ee(p.label),3)]),_:1},8,["items"]),m(),i(de,{href:`${t(k)("KUMA_DOCS_URL")}/policies/${s.value.path}/?${t(k)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"])]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"])]),m(),v.value!==null?(_(),P(se,{key:0,name:v.value,mesh:M.value,path:s.value.path,type:s.value.name},null,8,["name","mesh","path","type"])):N("",!0)])],2)):N("",!0)]),_:1})]}),_:1}))}});const Me=pe(ve,[["__scopeId","data-v-ed39ec55"]]);export{Me as default};
