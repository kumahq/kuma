import{l as Z,T as F,Z as W,V as j,v as G}from"./kongponents.es-dc880404.js";import{d as I,o as _,c as P,w as l,a as r,u as t,b as c,q as J,L as Y,r as d,m as b,s as H,f as X,e as y,g as x,J as L,N as V,t as ee,p as ae,j as te}from"./index-271b6183.js";import{_ as se}from"./PolicyDetails.vue_vue_type_script_setup_true_lang-b2c763d5.js";import{a as le,u as oe,b as ne,g as ie,f as re,e as ce,_ as pe}from"./RouteView.vue_vue_type_script_setup_true_lang-fadf0571.js";import{_ as ue}from"./RouteTitle.vue_vue_type_script_setup_true_lang-1f8fd421.js";import{D as me}from"./DataOverview-d4799d67.js";import{Q as $}from"./QueryParameter-70743f73.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-85609c66.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-9d9f8054.js";import"./ErrorBlock-2e363ab2.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-c91f8087.js";import"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-cfdfffdc.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ac8fa1e9.js";import"./TextWithCopyButton-3ab4305e.js";import"./toYaml-4e00099e.js";import"./TabsWidget-e239164d.js";import"./TagList-5e49a7a3.js";import"./StatusBadge-2721d158.js";const de=I({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(o){const a=o;return(k,S)=>(_(),P(t(F),{class:"docs-link",appearance:"outline",target:"_blank",to:a.href},{default:l(()=>[r(t(Z),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),c(`

    Documentation
  `)]),_:1},8,["to"]))}}),ye=o=>(ae("data-v-f1ec6f85"),o=o(),te(),o),fe={class:"kcard-stack"},he={class:"kcard-border"},_e=ye(()=>y("p",null,[y("strong",null,"Warning"),c(` This policy is experimental. If you encountered any problem please open an
                      `),y("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),ve=I({__name:"PolicyListView",props:{selectedPolicyName:{type:[String,null],required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(o){const a=o,k=le(),S=oe(),n=J(),B=Y(),i=ne(),{t:C}=ie(),T=d(!0),w=d(null),D=d(null),A=d(a.offset),v=d(a.selectedPolicyName),f=d({headers:[{label:"Name",key:"entity"},{label:"Type",key:"type"}],data:[]}),M=b(()=>n.params.mesh),s=b(()=>i.state.policyTypesByPath[a.policyPath]),U=b(()=>i.state.policyTypes.map(e=>({label:e.name,value:e.path,selected:e.path===a.policyPath}))),q=b(()=>i.state.policyTypes.filter(e=>(i.state.sidebar.insights.mesh.policies[e.name]??0)===0).map(e=>e.name));H(()=>n.params.mesh,function(){n.name===a.policyPath&&N(0)}),R();async function R(){const e=i.state.policyTypesByPath[a.policyPath];e!==void 0&&(await i.dispatch("updatePageTitle",""),await i.dispatch("updatePageTitle",e.name)),N(a.offset)}async function N(e){var g;A.value=e,$.set("offset",e>0?e:null),T.value=!0,w.value=null;const p=n.params.mesh,u=n.params.policyPath,m=L;try{const{items:h,next:Q}=await S.getAllPolicyEntitiesFromMesh({mesh:p,path:u},{size:m,offset:e});D.value=Q,f.value.data=z(h??[]),E({name:a.selectedPolicyName??((g=f.value.data[0])==null?void 0:g.entity.name)})}catch(h){f.value.data=[],h instanceof Error?w.value=h:console.error(h)}finally{T.value=!1}}function z(e){return e.map(p=>{const{type:u,name:m}=p,g={name:"policy-detail-view",params:{mesh:p.mesh,policyPath:n.params.policyPath,policy:m}};return{entity:p,detailViewRoute:g,type:u}})}function K(e){E({name:e.name})}function E({name:e}){v.value=e??null,$.set("policy",e??null)}function O(e){B.push({name:"policies-list-view",params:{...n.params,policyPath:e.value}})}return(e,p)=>(_(),P(ce,null,{default:l(()=>{var u;return[r(ue,{title:t(C)("policies.routes.items.title",{name:(u=s.value)==null?void 0:u.name})},null,8,["title"]),c(),r(re,null,{default:l(()=>[s.value?(_(),X("div",{key:0,class:V(["relative",s.value.path])},[y("div",fe,[y("div",he,[s.value.isExperimental?(_(),P(t(W),{key:0,"border-variant":"noBorder",class:"mb-4"},{body:l(()=>[r(t(j),{appearance:"warning"},{alertMessage:l(()=>[_e]),_:1})]),_:1})):x("",!0),c(),r(me,{"selected-entity-name":v.value??void 0,"page-size":t(L),error:w.value,"is-loading":T.value,"empty-state":{title:"No Data",message:`There are no ${s.value.name} policies present.`},"table-data":f.value,"table-data-is-empty":f.value.data.length===0,next:D.value,"page-offset":A.value,onTableAction:K,onLoadData:N},{additionalControls:l(()=>[r(t(G),{label:"Policies",items:U.value,"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:O},{"item-template":l(({item:m})=>[y("span",{class:V({"policy-type-empty":q.value.includes(m.label)})},ee(m.label),3)]),_:1},8,["items"]),c(),r(de,{href:`${t(k)("KUMA_DOCS_URL")}/policies/${s.value.path}/?${t(k)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"])]),default:l(()=>[c(`
              >
              `)]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"])]),c(),v.value!==null?(_(),P(se,{key:0,name:v.value,mesh:M.value,path:s.value.path,type:s.value.name},null,8,["name","mesh","path","type"])):x("",!0)])],2)):x("",!0)]),_:1})]}),_:1}))}});const Me=pe(ve,[["__scopeId","data-v-f1ec6f85"]]);export{Me as default};
