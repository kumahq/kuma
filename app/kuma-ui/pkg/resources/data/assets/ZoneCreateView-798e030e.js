import{d as L,v as _,x as de,S as me,r as b,o as l,j as p,m as t,h as i,w as r,g,i as e,az as X,aA as H,aB as pe,aC as fe,aD as _e,aE as j,l as n,R as V,a2 as O,e as W,f as K,D as o,P as ve,a3 as be,aF as U,F as q,k as y,G as ze,H as he,K as G,aG as ge}from"./index-ecc7df9d.js";import{_ as E}from"./CodeBlock.vue_vue_type_style_index_0_lang-1aa4056f.js";import{g as ke}from"./dataplane-0a086c06.js";const ye=["data-test-state"],Ee={class:"scanner-content"},Ce={class:"mr-1"},$e=L({__name:"EntityScanner",props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],setup(k,{emit:s}){const a=k,h=_(0),z=_(!1),C=_(!1),u=_(null);de(function(){$()}),me(function(){I()});function $(){z.value=!0,C.value=!1,I(),u.value=window.setInterval(async()=>{h.value++,await a.loaderFunction(),(h.value===a.retries||a.canComplete===!0)&&(I(),z.value=!1,C.value=!0,s("hide-siblings",!0))},a.interval)}function I(){u.value!==null&&window.clearInterval(u.value)}return(d,N)=>{const w=b("KEmptyState");return l(),p("div",{class:"scanner","data-test-state":z.value?"waiting":k.hasError?"error":"success"},[t("div",Ee,[i(w,{"cta-is-hidden":""},{title:r(()=>[t("span",Ce,[z.value?(l(),g(e(X),{key:0,color:e(H)},null,8,["color"])):k.hasError?(l(),g(e(pe),{key:1,color:e(fe)},null,8,["color"])):(l(),g(e(_e),{key:2,color:e(j)},null,8,["color"]))]),n(),z.value?V(d.$slots,"loading-title",{key:0}):k.hasError?V(d.$slots,"error-title",{key:1}):V(d.$slots,"complete-title",{key:2})]),message:r(()=>[z.value?V(d.$slots,"loading-content",{key:0}):k.hasError?V(d.$slots,"error-content",{key:1}):V(d.$slots,"complete-content",{key:2})]),_:3})])],8,ye)}}}),Ke={class:"form-step-title"},Ie=t("span",{class:"form-step-number"},"1",-1),we={class:"instruction-list"},Se={href:"https://helm.sh/docs/intro/install/"},Ve={class:"form-step-title"},Ne=t("span",{class:"form-step-number"},"2",-1),Te=t("p",null,"On your local machine, create a namespace in your Kubernetes cluster and pull down the kong Helm repo.",-1),Ze={class:"instruction-list"},qe={class:"form-step-title"},Le=t("span",{class:"form-step-number"},"3",-1),De={class:"form-step-title"},Ae=t("span",{class:"form-step-number"},"4",-1),Be={class:"field-group-label mt-4"},Re={class:"mt-4"},Ue=L({__name:"ZoneCreateKubernetesInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},zoneIngressEnabled:{type:Boolean,required:!0},zoneEgressEnabled:{type:Boolean,required:!0},token:{type:String,required:!0},base64EncodedToken:{type:String,required:!0}},setup(k){const s=k,a=O(),h=W(),z=K(()=>a.t("zones.form.kubernetes.secret.createSecretCommand",{token:s.base64EncodedToken}).trim()),C=K(()=>{const u={zoneName:s.zoneName,globalKdsAddress:s.globalKdsAddress,zoneIngressEnabled:String(s.zoneIngressEnabled),zoneEgressEnabled:String(s.zoneEgressEnabled)};return typeof h.params.virtualControlPlaneId=="string"&&(u.controlPlaneId=h.params.virtualControlPlaneId),a.t("zones.form.kubernetes.connectZone.config",u).trim()});return(u,$)=>(l(),p("div",null,[t("h3",Ke,[Ie,n(" "+o(e(a).t("zones.form.kubernetes.prerequisites.title")),1)]),n(),t("ul",we,[t("li",null,[t("b",null,o(e(a).t("zones.form.kubernetes.prerequisites.step1Label"))+o(s.zoneIngressEnabled?" "+e(a).t("zones.form.kubernetes.prerequisites.step1LabelAddendum"):""),1),n(`:
        `+o(e(a).t("zones.form.kubernetes.prerequisites.step1Description",{productName:e(a).t("common.product.name")})),1)]),n(),t("li",null,[t("b",null,o(e(a).t("zones.form.kubernetes.prerequisites.step2Label")),1),n(`:
        `+o(e(a).t("zones.form.kubernetes.prerequisites.step2Description")),1)]),n(),t("li",null,[t("a",Se,o(e(a).t("zones.form.kubernetes.prerequisites.step3LinkTitle")),1),n(" "+o(e(a).t("zones.form.kubernetes.prerequisites.step3Tail")),1)])]),n(),t("h3",Ve,[Ne,n(" "+o(e(a).t("zones.form.kubernetes.helm.title")),1)]),n(),Te,n(),t("ol",Ze,[t("li",null,[t("b",null,o(e(a).t("zones.form.kubernetes.helm.step1Description")),1),n(),i(E,{id:"zone-kubernetes-create-namespace",class:"mt-2",code:e(a).t("zones.form.kubernetes.helm.step1Command"),language:"bash"},null,8,["code"])]),n(),t("li",null,[t("b",null,o(e(a).t("zones.form.kubernetes.helm.step2Description")),1),n(),i(E,{id:"zone-kubernetes-add-charts-repo",class:"mt-2",code:e(a).t("zones.form.kubernetes.helm.step2Command"),language:"bash"},null,8,["code"])]),n(),t("li",null,[t("b",null,o(e(a).t("zones.form.kubernetes.helm.step3Description")),1),n(),i(E,{id:"zone-kubernetes-repo-update",class:"mt-2",code:e(a).t("zones.form.kubernetes.helm.step3Command"),language:"bash"},null,8,["code"])])]),n(),t("h3",qe,[Le,n(" "+o(e(a).t("zones.form.kubernetes.secret.title")),1)]),n(),t("p",null,o(e(a).t("zones.form.kubernetes.secret.createSecretDescription")),1),n(),i(E,{id:"zone-kubernetes-create-secret",class:"mt-4",code:z.value,language:"bash"},null,8,["code"]),n(),t("h3",De,[Ae,n(" "+o(e(a).t("zones.form.kubernetes.connectZone.title")),1)]),n(),t("p",null,o(e(a).t("zones.form.kubernetes.connectZone.configDescription")),1),n(),t("span",Be,o(e(a).t("zones.form.kubernetes.connectZone.configFileName")),1),n(),i(E,{id:"zone-kubernetes-config-code-block","data-testid":"zone-kubernetes-config",code:C.value,language:"yaml"},null,8,["code"]),n(),t("p",Re,o(e(a).t("zones.form.kubernetes.connectZone.connectDescription")),1),n(),i(E,{id:"zone-kubernetes-command-code-block",class:"mt-4",code:e(a).t("zones.form.kubernetes.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),Oe={class:"form-step-title"},xe=t("span",{class:"form-step-number"},"1",-1),Fe={class:"form-step-title"},Pe=t("span",{class:"form-step-number"},"2",-1),Me={class:"field-group-label mt-4"},Ge={class:"mt-4"},Xe=L({__name:"ZoneCreateUniversalInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},token:{type:String,required:!0}},setup(k){const s=k,a=O(),h=W(),z=K(()=>a.t("zones.form.universal.saveToken.saveTokenCommand",{token:s.token}).trim()),C=K(()=>{const u={zoneName:s.zoneName,globalKdsAddress:s.globalKdsAddress};return typeof h.params.virtualControlPlaneId=="string"&&(u.controlPlaneId=h.params.virtualControlPlaneId),a.t("zones.form.universal.connectZone.config",u).trim()});return(u,$)=>(l(),p("div",null,[t("h3",Oe,[xe,n(" "+o(e(a).t("zones.form.universal.saveToken.title")),1)]),n(),t("p",null,o(e(a).t("zones.form.universal.saveToken.saveTokenDescription")),1),n(),i(E,{id:"zone-kubernetes-token",class:"mt-4",code:z.value,language:"bash"},null,8,["code"]),n(),t("h3",Fe,[Pe,n(" "+o(e(a).t("zones.form.universal.connectZone.title")),1)]),n(),t("p",null,o(e(a).t("zones.form.universal.connectZone.configDescription")),1),n(),t("span",Me,o(e(a).t("zones.form.universal.connectZone.configFileName")),1),n(),i(E,{id:"zone-universal-config-code-block","data-testid":"zone-universal-config",class:"mt-4",code:C.value,language:"yaml"},null,8,["code"]),n(),t("p",Ge,o(e(a).t("zones.form.universal.connectZone.connectDescription")),1),n(),i(E,{id:"zone-universal-connect-command-code-block",class:"mt-4",code:e(a).t("zones.form.universal.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),He={class:"form-wrapper"},je={key:1},We={key:2},Je={class:"form"},Qe={class:"form-header"},Ye={class:"form-title"},en={class:"text-gradient"},nn={key:0},tn={key:0},on={class:"fact-list"},sn={class:"form-section"},an={class:"form-section__header"},rn={class:"form-section-title"},ln={key:0},cn={class:"form-section__content"},un={class:"form-section","data-testid":"connect-zone-instructions"},dn={class:"form-section__header"},mn={class:"form-section-title"},pn={key:0},fn={class:"form-section__content"},_n={class:"field-group-list"},vn={class:"field-group"},bn={class:"field-group-label"},zn={class:"radio-button-group"},hn={class:"field-group"},gn={class:"field-group-label"},kn={class:"radio-button-group"},yn={class:"field-group"},En={class:"field-group-label"},Cn={class:"radio-button-group"},$n={class:"form-section"},Kn={class:"form-section__header"},In={class:"form-section-title"},wn={key:0},Sn={class:"form-section__content"},Vn={class:"form-section"},Nn={class:"form-section__header"},Tn={class:"form-section-title"},Zn={key:0},qn={class:"form-section__content"},Ln={class:"mt-2"},Rn=L({__name:"ZoneCreateView",setup(k){const{t:s,tm:a}=O(),h=ve(),z=be(),C=/^(?![-0-9])[a-z0-9-]{1,63}$/,u=_(null),$=_(!1),I=_(!1),d=_(null),N=_(null),w=_(!1),D=_(null),v=_(""),S=_("kubernetes"),A=_(!0),B=_(!0),T=K(()=>u.value!==null&&u.value.token?u.value.token:""),J=K(()=>T.value!==""?window.btoa(T.value):""),Q=K(()=>v.value===""||$.value||u.value!==null),R=K(()=>{if(N.value!==null)return N.value;if(d.value instanceof U){const f=d.value.invalidParameters.find(c=>c.field==="name");if(f!==void 0)return f.reason}return null});async function Y(){$.value=!0,d.value=null;try{if(!x(v.value))return;u.value=await h.createZone({name:v.value})}catch(f){f instanceof Error?d.value=f:console.error(f)}finally{$.value=!1}}function x(f){const c=C.test(f);return c?N.value=null:N.value=s("zones.create.invalidNameError"),c}async function ee(){w.value=!1,D.value=null;try{const f=await h.getZoneOverview({name:v.value}),c=ke(f.zoneInsight);w.value=c==="online"}catch(f){f instanceof Error?D.value=f:console.error(f)}}function F(){I.value=!I.value}return(f,c)=>{const ne=b("RouteTitle"),Z=b("KButton"),te=b("KAlert"),oe=b("KLabel"),se=b("KInput"),P=b("KRadio"),M=b("KInputSwitch"),ae=b("DataSource"),re=b("KCard"),le=b("KModal"),ie=b("AppView"),ce=b("RouteView");return l(),g(ce,{name:"zone-create-view",attrs:{class:"is-fullscreen"}},{default:r(()=>[i(ie,{fullscreen:!0,breadcrumbs:[]},{title:r(()=>[t("h1",null,[i(ne,{title:e(s)("zones.routes.create.title"),render:!0},null,8,["title"])])]),actions:r(()=>[T.value===""||w.value?(l(),g(Z,{key:0,appearance:"outline","data-testid":"exit-button",to:{name:"zone-cp-list-view"}},{default:r(()=>[n(o(e(s)("zones.form.exit")),1)]),_:1})):(l(),g(Z,{key:1,appearance:"outline","data-testid":"exit-button",onClick:F},{default:r(()=>[n(o(e(s)("zones.form.exit")),1)]),_:1}))]),default:r(()=>[n(),n(),t("div",He,[d.value!==null?(l(),g(te,{key:0,appearance:"danger",class:"mb-4","dismiss-type":"icon","data-testid":"create-zone-error"},{alertMessage:r(()=>[d.value instanceof e(U)&&[409,500].includes(d.value.status)?(l(),p(q,{key:0},[t("p",null,o(e(s)(`zones.create.status_error.${d.value.status}.title`,{name:v.value})),1),n(),t("p",null,o(e(s)(`zones.create.status_error.${d.value.status}.description`)),1)],64)):d.value instanceof e(U)?(l(),p("p",je,o(e(s)("common.error_state.api_error",{status:d.value.status,title:d.value.title})),1)):(l(),p("p",We,o(e(s)("common.error_state.default_error")),1))]),_:1})):y("",!0),n(),i(re,{class:"form-card"},{body:r(()=>[t("div",Je,[t("div",Qe,[t("div",null,[t("h1",Ye,[t("span",en,o(e(s)("zones.form.title")),1)]),n(),e(s)("zones.form.description")!==" "?(l(),p("p",nn,o(e(s)("zones.form.description")),1)):y("",!0)]),n(),e(a)("zones.form.facts").length>0?(l(),p("div",tn,[t("ul",on,[(l(!0),p(q,null,ze(e(a)("zones.form.facts"),(m,ue)=>(l(),p("li",{key:ue,class:"fact-list__item"},[i(e(he),{color:e(j)},null,8,["color"]),n(" "+o(m),1)]))),128))])])):y("",!0)]),n(),t("div",sn,[t("div",an,[t("h2",rn,o(e(s)("zones.form.section.name.title")),1),n(),e(s)("zones.form.section.name.description")!==" "?(l(),p("p",ln,o(e(s)("zones.form.section.name.description")),1)):y("",!0)]),n(),t("div",cn,[t("div",null,[i(oe,{for:"zone-name",required:"","tooltip-attributes":{placement:"right"}},{tooltip:r(()=>[n(o(e(s)("zones.form.name_tooltip")),1)]),default:r(()=>[n(o(e(s)("zones.form.nameLabel"))+" ",1)]),_:1}),n(),i(se,{id:"zone-name",modelValue:v.value,"onUpdate:modelValue":c[0]||(c[0]=m=>v.value=m),type:"text",name:"zone-name","data-testid":"name-input","data-test-error-type":R.value!==null?"invalid-dns-name":void 0,"has-error":R.value!==null,"error-message":R.value??void 0,disabled:u.value!==null,onBlur:c[1]||(c[1]=m=>x(v.value))},null,8,["modelValue","data-test-error-type","has-error","error-message","disabled"])]),n(),i(Z,{appearance:"primary",class:"mt-4",disabled:Q.value,"data-testid":"create-zone-button",onClick:Y},{default:r(()=>[$.value?(l(),g(e(X),{key:0,color:e(H),size:e(G)},null,8,["color","size"])):(l(),g(e(ge),{key:1,size:e(G)},null,8,["size"])),n(" "+o(e(s)("zones.form.createZoneButtonLabel")),1)]),_:1},8,["disabled"])])]),n(),u.value!==null?(l(),p(q,{key:0},[t("div",un,[t("div",dn,[t("h2",mn,o(e(s)("zones.form.section.configuration.title")),1),n(),e(s)("zones.form.section.configuration.description")!==" "?(l(),p("p",pn,o(e(s)("zones.form.section.configuration.description")),1)):y("",!0)]),n(),t("div",fn,[t("div",_n,[t("div",vn,[t("span",bn,o(e(s)("zones.form.environmentLabel"))+` *
                        `,1),n(),t("div",zn,[i(P,{id:"zone-environment-universal",modelValue:S.value,"onUpdate:modelValue":c[2]||(c[2]=m=>S.value=m),"selected-value":"universal",name:"zone-environment","data-testid":"environment-universal-radio-button"},{default:r(()=>[n(o(e(s)("zones.form.universalLabel")),1)]),_:1},8,["modelValue"]),n(),i(P,{id:"zone-environment-kubernetes",modelValue:S.value,"onUpdate:modelValue":c[3]||(c[3]=m=>S.value=m),"selected-value":"kubernetes",name:"zone-environment","data-testid":"environment-kubernetes-radio-button"},{default:r(()=>[n(o(e(s)("zones.form.kubernetesLabel")),1)]),_:1},8,["modelValue"])])]),n(),S.value==="kubernetes"?(l(),p(q,{key:0},[t("div",hn,[t("span",gn,o(e(s)("zones.form.zoneIngressLabel"))+` *
                          `,1),n(),t("div",kn,[i(M,{id:"zone-ingress-enabled",modelValue:A.value,"onUpdate:modelValue":c[4]||(c[4]=m=>A.value=m),"data-testid":"ingress-input-switch"},{label:r(()=>[n(o(e(s)("zones.form.zoneIngressEnabledLabel")),1)]),_:1},8,["modelValue"])])]),n(),t("div",yn,[t("span",En,o(e(s)("zones.form.zoneEgressLabel"))+` *
                          `,1),n(),t("div",Cn,[i(M,{id:"zone-egress-enabled",modelValue:B.value,"onUpdate:modelValue":c[5]||(c[5]=m=>B.value=m),"data-testid":"egress-input-switch"},{label:r(()=>[n(o(e(s)("zones.form.zoneEgressEnabledLabel")),1)]),_:1},8,["modelValue"])])])],64)):y("",!0)])])]),n(),t("div",$n,[t("div",Kn,[t("h2",In,o(e(s)("zones.form.section.connect_zone.title")),1),n(),e(s)("zones.form.section.connect_zone.description")!==" "?(l(),p("p",wn,o(e(s)("zones.form.section.connect_zone.description")),1)):y("",!0)]),n(),t("div",Sn,[i(ae,{src:"/control-plane/addresses"},{default:r(({data:m})=>[typeof m<"u"?(l(),p(q,{key:0},[S.value==="universal"?(l(),g(Xe,{key:0,"zone-name":v.value,token:T.value,"global-kds-address":m.kds},null,8,["zone-name","token","global-kds-address"])):(l(),g(Ue,{key:1,"zone-name":v.value,"zone-ingress-enabled":A.value,"zone-egress-enabled":B.value,token:T.value,"base64-encoded-token":J.value,"global-kds-address":m.kds},null,8,["zone-name","zone-ingress-enabled","zone-egress-enabled","token","base64-encoded-token","global-kds-address"]))],64)):y("",!0)]),_:1})])]),n(),t("div",Vn,[t("div",Nn,[t("h2",Tn,o(e(s)("zones.form.section.scanner.title")),1),n(),e(s)("zones.form.section.scanner.description")!==" "?(l(),p("p",Zn,o(e(s)("zones.form.section.scanner.description")),1)):y("",!0)]),n(),t("div",qn,[i($e,{"loader-function":ee,"has-error":D.value!==null,"can-complete":w.value,"data-testid":"zone-connected-scanner"},{"loading-title":r(()=>[n(o(e(s)("zones.form.scan.waitTitle")),1)]),"complete-title":r(()=>[n(o(e(s)("zones.form.scan.completeTitle")),1)]),"complete-content":r(()=>[t("p",null,o(e(s)("zones.form.scan.completeDescription",{name:v.value})),1),n(),t("p",Ln,[i(Z,{appearance:"primary",to:{name:"zone-cp-detail-view",params:{zone:v.value}}},{default:r(()=>[n(o(e(s)("zones.form.scan.completeButtonLabel",{name:v.value})),1)]),_:1},8,["to"])])]),"error-title":r(()=>[t("h3",null,o(e(s)("zones.form.scan.errorTitle")),1)]),"error-content":r(()=>[t("p",null,o(e(s)("zones.form.scan.errorDescription")),1)]),_:1},8,["has-error","can-complete"])])])],64)):y("",!0)])]),_:1})]),n(),i(le,{"is-visible":I.value,title:e(s)("zones.form.confirm_modal.title"),"data-testid":"confirm-exit-modal",onCanceled:F,onProceed:c[6]||(c[6]=m=>e(z).push({name:"zone-cp-list-view"}))},{"header-content":r(()=>[n(o(e(s)("zones.form.confirm_modal.title")),1)]),"body-content":r(()=>[n(o(e(s)("zones.form.confirm_modal.body")),1)]),"action-buttons":r(()=>[i(Z,{appearance:"primary",to:{name:"zone-cp-list-view"},"data-testid":"confirm-exit-button"},{default:r(()=>[n(o(e(s)("zones.form.confirm_modal.action_button")),1)]),_:1})]),_:1},8,["is-visible","title"])]),_:1})]),_:1})}}});export{Rn as default};
