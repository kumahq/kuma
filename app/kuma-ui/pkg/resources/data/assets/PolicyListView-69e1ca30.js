import{d as H,o as p,k as L,w as n,a as f,u as a,h as Z,b as o,E as K,l as J,C as ee,ai as ae,m as te,i as se,r as l,j as R,n as F,P as M,t as le,H as ne,c as $,X as oe,x as E,aj as ie,f as i,y as x,a9 as B,F as re,z as ue,L as ce,N as pe,_ as me}from"./index-c8ce0213.js";import{_ as de}from"./PolicyConnections.vue_vue_type_script_setup_true_lang-713a6309.js";import{D as ye}from"./DataOverview-8df7b6a3.js";import{F as ve}from"./FrameSkeleton-6fbe1de7.js";import{_ as fe}from"./LabelList.vue_vue_type_style_index_0_lang-8d53c57d.js";import{T as he}from"./TabsWidget-40ac9857.js";import{_ as _e}from"./YamlView.vue_vue_type_script_setup_true_lang-b7cb6ad2.js";import{Q as O}from"./QueryParameter-70743f73.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-ec77a77d.js";import"./ErrorBlock-a8c0484c.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-171e2abf.js";import"./TagList-64dd4a55.js";import"./StatusBadge-584f994a.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ad640f64.js";import"./toYaml-4e00099e.js";const ge=H({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(m){const r=m;return(T,q)=>(p(),L(a(K),{class:"docs-link",appearance:"outline",target:"_blank",to:r.href},{default:n(()=>[f(a(Z),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),o(`

    Documentation
  `)]),_:1},8,["to"]))}}),be=m=>(ce("data-v-38386bed"),m=m(),pe(),m),Pe={key:0,class:"mb-4"},ke=be(()=>i("p",null,[i("strong",null,"Warning"),o(` This policy is experimental. If you encountered any problem please open an
            `),i("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),we={class:"entity-heading","data-testid":"policy-single-entity"},Ee={"data-testid":"policy-overview-tab"},Se={class:"config-wrapper"},xe=H({__name:"PolicyListView",props:{selectedPolicyName:{type:String,required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(m){const r=m,T=J(),q=ee(),Q=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],W=ae(),h=te(),A=se(),_=l(!0),g=l(!1),d=l(null),y=l(!0),v=l(!1),b=l(!1),S=l(!1),P=l({}),k=l(null),C=l(null),U=l(r.offset),V=l({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]}),u=R(()=>A.state.policyTypesByPath[r.policyPath]),j=R(()=>A.state.policyTypes.map(e=>({label:e.name,value:e.path,selected:e.path===r.policyPath}))),G=R(()=>A.state.policyTypes.filter(e=>(A.state.sidebar.insights.mesh.policies[e.name]??0)===0).map(e=>e.name));F(()=>h.params.mesh,function(){h.name===r.policyPath&&(_.value=!0,g.value=!1,y.value=!0,v.value=!1,b.value=!1,S.value=!1,d.value=null,D(0))}),F(()=>h.query.ns,function(){_.value=!0,g.value=!1,y.value=!0,v.value=!1,b.value=!1,S.value=!1,d.value=null,D(0)}),D(r.offset);async function D(e){U.value=e,O.set("offset",e>0?e:null),_.value=!0,d.value=null;const s=h.query.ns||null,t=h.params.mesh,w=u.value.path;try{let c;if(t!==null&&s!==null)c=[await T.getSinglePolicyEntity({mesh:t,path:w,name:s})],C.value=null;else{const N={size:M,offset:e},I=await T.getAllPolicyEntitiesFromMesh({mesh:t,path:w},N);c=I.items??[],C.value=I.next}if(c.length>0){V.value.data=c.map(I=>Y(I)),S.value=!1,g.value=!1;const N=r.selectedPolicyName??c[0].name;await z({name:N,mesh:t,path:w})}else V.value.data=[],S.value=!0,g.value=!0,v.value=!0}catch(c){c instanceof Error?d.value=c:console.error(c),g.value=!0}finally{_.value=!1,y.value=!1}}function X(e){W.push({name:"policy",params:{...h.params,policyPath:e.value}})}function Y(e){if(!e.mesh)return e;const s=e,t={name:"mesh-detail-view",params:{mesh:e.mesh}};return s.meshRoute=t,s}async function z(e){b.value=!1,y.value=!0,v.value=!1;try{const s=await T.getSinglePolicyEntity({mesh:e.mesh,path:u.value.path,name:e.name});if(s){const t=["type","name","mesh"];P.value=le(s,t),O.set("policy",P.value.name),k.value=ne(s)}else P.value={},v.value=!0}catch(s){b.value=!0,console.error(s)}finally{y.value=!1}}return(e,s)=>a(u)?(p(),$("div",{key:0,class:B(["relative",a(u).path])},[a(u).isExperimental?(p(),$("div",Pe,[f(a(oe),{appearance:"warning"},{alertMessage:n(()=>[ke]),_:1})])):E("",!0),o(),f(ve,null,{default:n(()=>[f(ye,{"selected-entity-name":P.value.name,"page-size":a(M),error:d.value,"is-loading":_.value,"empty-state":{title:"No Data",message:`There are no ${a(u).name} policies present.`},"table-data":V.value,"table-data-is-empty":S.value,next:C.value,"page-offset":U.value,onTableAction:z,onLoadData:D},{additionalControls:n(()=>[f(a(ie),{label:"Policies",items:a(j),"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:X},{"item-template":n(({item:t})=>[i("span",{class:B({"policy-type-empty":a(G).includes(t.label)})},x(t.label),3)]),_:1},8,["items"]),o(),f(ge,{href:`${a(q)("KUMA_DOCS_URL")}/policies/${a(u).path}/?${a(q)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"]),o(),e.$route.query.ns?(p(),L(a(K),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"policy",params:{policyPath:r.policyPath}}},{default:n(()=>[o(`
            View all
          `)]),_:1},8,["to"])):E("",!0)]),default:n(()=>[o(`
        >
        `)]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"]),o(),g.value===!1?(p(),L(he,{key:0,"has-error":d.value!==null,error:d.value,"is-loading":_.value,tabs:Q},{tabHeader:n(()=>[i("h1",we,x(a(u).name)+": "+x(P.value.name),1)]),overview:n(()=>[f(fe,{"has-error":b.value,"is-loading":y.value,"is-empty":v.value},{default:n(()=>[i("div",Ee,[i("ul",null,[(p(!0),$(re,null,ue(P.value,(t,w)=>(p(),$("li",{key:w},[i("h4",null,x(w),1),o(),i("p",null,x(t),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),o(),i("div",Se,[k.value!==null?(p(),L(_e,{key:0,id:"code-block-policy","has-error":b.value,"is-loading":y.value,"is-empty":v.value,content:k.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):E("",!0)])]),"affected-dpps":n(()=>[k.value!==null?(p(),L(de,{key:0,mesh:k.value.mesh,"policy-name":k.value.name,"policy-type":a(u).path},null,8,["mesh","policy-name","policy-type"])):E("",!0)]),_:1},8,["has-error","error","is-loading"])):E("",!0)]),_:1})],2)):E("",!0)}});const Be=me(xe,[["__scopeId","data-v-38386bed"]]);export{Be as default};
